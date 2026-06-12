package controller

import (
	"net/http"

	"github.com/CatMsg/Nova_x-ui/v3/web/entity"
	"github.com/CatMsg/Nova_x-ui/v3/web/middleware"
	"github.com/CatMsg/Nova_x-ui/v3/web/session"

	"github.com/gin-gonic/gin"
)

// PanelController is the main controller for the Nova_x-ui panel, managing sub-controllers.
type PanelController struct {
	BaseController

	settingController     *SettingController
	xraySettingController *XraySettingController
}

// NewPanelController creates a new PanelController and initializes its routes.
func NewPanelController(g *gin.RouterGroup) *PanelController {
	a := &PanelController{}
	a.initRouter(g)
	return a
}

// initRouter sets up the main panel routes and initializes sub-controllers.
func (a *PanelController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/panel")
	g.Use(a.checkLogin)
	g.Use(middleware.CSRFMiddleware())

	g.GET("/", a.index)
	g.GET("/inbounds", a.inbounds)
	g.GET("/nodes", a.nodes)
	g.GET("/settings", a.settings)
	g.GET("/xray", a.xraySettings)
	g.GET("/api-docs", a.apiDocs)

	// SPA pages built by Vite don't have a server-rendered <meta name="csrf-token">,
	// so they fetch the session token via this endpoint at startup and replay it
	// on subsequent unsafe requests through axios.
	g.GET("/csrf-token", a.csrfToken)

	a.settingController = NewSettingController(g)
	a.xraySettingController = NewXraySettingController(g)
}

// All four panel pages now serve the Vue 3 builds from web/dist/
// instead of rendering the previous Go templates. Each handler is a
// thin wrapper around serveDistPage so the basePath injection +
// no-cache headers stay centralised.

// index renders the main panel index page.
func (a *PanelController) index(c *gin.Context) {
	serveDistPage(c, "index.html")
}

// inbounds renders the inbounds management page.
func (a *PanelController) inbounds(c *gin.Context) {
	serveDistPage(c, "inbounds.html")
}

// nodes renders the multi-panel nodes management page.
func (a *PanelController) nodes(c *gin.Context) {
	serveDistPage(c, "nodes.html")
}

// settings renders the settings management page.
func (a *PanelController) settings(c *gin.Context) {
	serveDistPage(c, "settings.html")
}

// xraySettings renders the Xray settings page.
func (a *PanelController) xraySettings(c *gin.Context) {
	serveDistPage(c, "xray.html")
}

// apiDocs renders the in-panel API documentation page.
func (a *PanelController) apiDocs(c *gin.Context) {
	serveDistPage(c, "api-docs.html")
}

// csrfToken returns the session CSRF token to authenticated SPA clients.
// The endpoint is GET (a safe method) so it bypasses CSRFMiddleware itself,
// but checkLogin still gates the response — anonymous callers get 401/redirect.
func (a *PanelController) csrfToken(c *gin.Context) {
	token, err := session.EnsureCSRFToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Obj: token})
}
