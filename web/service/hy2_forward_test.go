package service

import (
	"encoding/json"
	"testing"

	"github.com/CatMsg/Nova_x-ui/v3/database/model"
)

func TestGetHy2ServerPorts(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []int
		wantErr bool
	}{
		{
			name: "nested array with duplicates and junk",
			raw:  `{"hysteriaSettings":{"server_ports":["2026","2027","2026","junk",0,65536]}}`,
			want: []int{2026, 2027},
		},
		{
			name: "nested comma string",
			raw:  `{"hysteriaSettings":{"server_ports":"2026, 2028, 2028, 2030"}}`,
			want: []int{2026, 2028, 2030},
		},
		{
			name: "top level fallback",
			raw:  `{"server_ports":[2026,2027,2027]}`,
			want: []int{2026, 2027},
		},
		{
			name:    "invalid json",
			raw:     `{`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHy2ServerPorts(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("getHy2ServerPorts() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len(got) = %d, want %d, got=%v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got[%d] = %d, want %d, full=%v", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

func TestIsHy2Inbound(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "protocol hysteria2",
			in:   `{"protocol":"hysteria2","settings":"{}"}`,
			want: true,
		},
		{
			name: "protocol hysteria with version 2",
			in:   `{"protocol":"hysteria","settings":"{\"version\":2}"}`,
			want: true,
		},
		{
			name: "protocol hysteria version 1",
			in:   `{"protocol":"hysteria","settings":"{\"version\":1}"}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inbound struct {
				Protocol string `json:"protocol"`
				Settings string `json:"settings"`
			}
			if err := json.Unmarshal([]byte(tt.in), &inbound); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			got := isHy2Inbound(&model.Inbound{
				Protocol: model.Protocol(inbound.Protocol),
				Settings: inbound.Settings,
			})
			if got != tt.want {
				t.Fatalf("isHy2Inbound() = %v, want %v", got, tt.want)
			}
		})
	}
}
