#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import os
import platform
import shutil
import stat
import sys
import tempfile
import urllib.request
import zipfile
from pathlib import Path


REPO = "XTLS/Xray-core"
DEFAULT_VERSION = "v26.3.27"
LATEST_RELEASE_API = f"https://api.github.com/repos/{REPO}/releases/latest"
TAG_RELEASE_API = f"https://api.github.com/repos/{REPO}/releases/tags/{{tag}}"
USER_AGENT = "Nova_x-ui-xray-provisioner"


def request_json(url: str) -> dict:
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT, "Accept": "application/vnd.github+json"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.load(resp)


def release_info(version: str) -> dict:
    if version == "latest":
        return request_json(LATEST_RELEASE_API)
    return request_json(TAG_RELEASE_API.format(tag=version))


def current_platform_asset_name() -> str | None:
    goos = sys.platform
    machine = platform.machine().lower()

    if goos == "darwin":
        if machine in {"arm64", "aarch64"}:
            return "Xray-macos-arm64-v8a.zip"
        if machine in {"x86_64", "amd64", "i386", "i686"}:
            return "Xray-macos-64.zip"
    elif goos.startswith("linux"):
        if machine in {"arm64", "aarch64"}:
            return "Xray-linux-arm64-v8a.zip"
        if machine in {"x86_64", "amd64"}:
            return "Xray-linux-64.zip"
        if machine in {"i386", "i686"}:
            return "Xray-linux-32.zip"
    elif goos in {"win32", "cygwin", "msys"}:
        if machine in {"arm64", "aarch64"}:
            return "Xray-windows-arm64-v8a.zip"
        if machine in {"x86_64", "amd64"}:
            return "Xray-windows-64.zip"
        if machine in {"i386", "i686"}:
            return "Xray-windows-32.zip"
    elif goos.startswith("freebsd"):
        if machine in {"arm64", "aarch64"}:
            return "Xray-freebsd-arm64-v8a.zip"
        if machine in {"arm", "armv7l"}:
            return "Xray-freebsd-arm32-v7a.zip"
        if machine in {"x86_64", "amd64"}:
            return "Xray-freebsd-64.zip"
        if machine in {"i386", "i686"}:
            return "Xray-freebsd-32.zip"
    elif goos.startswith("openbsd"):
        if machine in {"arm64", "aarch64"}:
            return "Xray-openbsd-arm64-v8a.zip"
        if machine in {"arm", "armv7l"}:
            return "Xray-openbsd-arm32-v7a.zip"
        if machine in {"x86_64", "amd64"}:
            return "Xray-openbsd-64.zip"
        if machine in {"i386", "i686"}:
            return "Xray-openbsd-32.zip"
    return None


def current_goarch() -> str:
    machine = platform.machine().lower()
    if machine in {"x86_64", "amd64"}:
        return "amd64"
    if machine in {"i386", "i686"}:
        return "386"
    if machine in {"arm64", "aarch64"}:
        return "arm64"
    if machine in {"arm", "armv7l"}:
        return "arm"
    return machine


def get_cache_root() -> Path:
    override = os.environ.get("XUI_CACHE_ROOT")
    if override:
        return Path(override).expanduser()
    if sys.platform == "darwin":
        return Path.home() / "Library" / "Caches" / "Nova_x-ui"
    xdg_cache = os.environ.get("XDG_CACHE_HOME")
    if xdg_cache:
        return Path(xdg_cache).expanduser() / "Nova_x-ui"
    return Path.home() / ".cache" / "Nova_x-ui"


def download_file(url: str, dest: Path) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    tmp = dest.with_suffix(dest.suffix + ".tmp")
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT, "Accept": "application/octet-stream"})
    with urllib.request.urlopen(req, timeout=120) as resp, tmp.open("wb") as fh:
        shutil.copyfileobj(resp, fh)
    tmp.replace(dest)


def link_or_copy(source: Path, dest: Path) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    if dest.exists() or dest.is_symlink():
        dest.unlink()
    if os.name == "nt":
        shutil.copy2(source, dest)
        return
    dest.symlink_to(source)


def install_local_xray(bin_dir: Path, install_dir: Path, link_name: str) -> bool:
    xray_path = shutil.which("xray")
    if not xray_path:
        return False

    source_binary = Path(xray_path)
    if not source_binary.exists():
        return False

    bin_dir_resolved = bin_dir.resolve()
    try:
        source_resolved = source_binary.resolve()
        if bin_dir_resolved in source_resolved.parents:
            return False
    except OSError:
        pass

    source_dir = source_binary.parent
    geoip_source = source_dir / "geoip.dat"
    geosite_source = source_dir / "geosite.dat"
    if not geoip_source.exists() or not geosite_source.exists():
        return False

    install_dir.mkdir(parents=True, exist_ok=True)
    install_binary = install_dir / source_binary.name
    shutil.copy2(source_binary, install_binary)
    install_binary.chmod(install_binary.stat().st_mode | stat.S_IEXEC)

    for name in ("geoip.dat", "geosite.dat"):
        source = source_dir / name
        if source.exists():
            shutil.copy2(source, install_dir / name)

    link_or_copy(install_binary, bin_dir / link_name)
    for name in ("geoip.dat", "geosite.dat"):
        source = install_dir / name
        if source.exists():
            link_or_copy(source, bin_dir / name)
    return True


def ensure_current_binary(release: dict, bin_dir: Path, cache_dir: Path, asset_name: str) -> None:
    assets = {asset["name"]: asset["browser_download_url"] for asset in release.get("assets", []) if asset["name"].endswith(".zip")}
    if asset_name not in assets:
        available = ", ".join(sorted(assets))
        raise SystemExit(f"missing Xray asset {asset_name}; available: {available}")

    version = release["tag_name"]
    install_dir = cache_dir / "install" / version / asset_name.removesuffix(".zip")
    install_binary_name = "xray.exe" if os.name == "nt" else "xray"
    install_binary = install_dir / install_binary_name
    target_link = bin_dir / f"xray-{platform.system().lower()}-{current_goarch()}"

    if target_link.exists() and target_link.is_symlink():
        try:
            if target_link.resolve(strict=False) == install_binary.resolve(strict=False) and install_binary.exists():
                geoip_link = bin_dir / "geoip.dat"
                geosite_link = bin_dir / "geosite.dat"
                if geoip_link.exists() and geosite_link.exists():
                    return
        except OSError:
            pass

    if install_local_xray(bin_dir, install_dir, target_link.name):
        print(f"installed {target_link} from local xray")
        return

    zip_path = cache_dir / version / asset_name
    if not zip_path.exists():
        print(f"downloading {asset_name} for {version}")
        download_file(assets[asset_name], zip_path)
    else:
        print(f"using cached {asset_name} for {version}")

    with tempfile.TemporaryDirectory(prefix="xray-provision-") as tmpdir:
        extract_dir = Path(tmpdir) / "extract"
        extract_dir.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(zip_path) as zf:
            zf.extractall(extract_dir)

        candidates = [p for p in extract_dir.iterdir() if p.is_file() and p.name in {"xray", "xray.exe"}]
        if not candidates:
            raise SystemExit(f"archive {zip_path} does not contain xray binary")
        source_binary = candidates[0]

        install_dir.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source_binary, install_binary)
        install_binary.chmod(install_binary.stat().st_mode | stat.S_IEXEC)

        for name in ("geoip.dat", "geosite.dat"):
            source = extract_dir / name
            if source.exists():
                shutil.copy2(source, install_dir / name)

        link_or_copy(install_binary, target_link)
        for name in ("geoip.dat", "geosite.dat"):
            source = install_dir / name
            if source.exists():
                link_or_copy(source, bin_dir / name)

        print(f"installed {target_link}")


def prefetch_all_assets(release: dict, cache_dir: Path) -> None:
    version = release["tag_name"]
    assets = sorted(
        asset["name"]
        for asset in release.get("assets", [])
        if asset["name"].endswith(".zip")
    )
    asset_urls = {asset["name"]: asset["browser_download_url"] for asset in release.get("assets", []) if asset["name"].endswith(".zip")}

    for asset_name in assets:
        zip_path = cache_dir / version / asset_name
        if zip_path.exists():
            print(f"cached {asset_name}")
            continue
        print(f"prefetching {asset_name}")
        download_file(asset_urls[asset_name], zip_path)


def main() -> int:
    parser = argparse.ArgumentParser(description="Provision Xray binaries for local Nova_x-ui runs.")
    parser.add_argument("--all", action="store_true", help="download all official Xray zip assets into the local cache")
    parser.add_argument("--version", default=os.environ.get("XRAY_VERSION", DEFAULT_VERSION), help="Xray release tag or latest")
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[1]
    bin_dir = Path(os.environ.get("XUI_BIN_FOLDER", root / "bin"))
    cache_dir = get_cache_root() / "xray"

    release = release_info(args.version)
    asset_name = current_platform_asset_name()
    if not asset_name:
        raise SystemExit(f"unsupported current platform: {sys.platform}/{platform.machine()}")

    if args.all or os.environ.get("XUI_DOWNLOAD_XRAY_ALL", "0") == "1":
        prefetch_all_assets(release, cache_dir)

    ensure_current_binary(release, bin_dir, cache_dir, asset_name)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
