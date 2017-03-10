package controller

import (
	"strconv"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/goadesign/goa"
	errs "github.com/pkg/errors"
)

const (
	sourceLinkTypesRouteEnd = "/source-link-types"
	targetLinkTypesRouteEnd = "/target-link-types"
)

// WorkitemtypeController implements the workitemtype resource.
type WorkitemtypeController struct {
	*goa.Controller
	db            application.DB
	configuration cacheControlConfiguration
}

type cacheControlConfiguration interface {
	GetWorkItemTypeCacheControlMaxAge() string
}

// NewWorkitemtypeController creates a workitemtype controller.
func NewWorkitemtypeController(service *goa.Service, db application.DB, configuration cacheControlConfiguration) *WorkitemtypeController {
	return &WorkitemtypeController{
		Controller:    service.NewController("WorkitemtypeController"),
		db:            db,
		configuration: configuration,
	}
}

// Show runs the show action.
func (c *WorkitemtypeController) Show(ctx *app.ShowWorkitemtypeContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		res, err := appl.WorkItemTypes().Load(ctx.Context, ctx.WitID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		// check the "If-Modified-Since header against the last update timestamp"
		// HTTP header does not include microseconds, so we need to ignore them in the "updated_at" record field.
		updatedAt := res.Data.Attributes.UpdatedAt.Truncate(time.Second)
		if ctx.IfModifiedSince != nil {
			logrus.Debug(nil, map[string]interface{}{
				IfModifiedSince: ctx.IfModifiedSince.UTC(),
				LastModified:    updatedAt,
			}, "work item type conditional query")

			if ctx.IfModifiedSince != nil && ctx.IfModifiedSince.UTC().After(updatedAt) {
				return ctx.NotModified()
			}
		}
		// check the ETag
		etag := strconv.Itoa(res.Data.Attributes.Version)
		if ctx.IfNoneMatch != nil {
			logrus.Debug(nil, map[string]interface{}{
				IfNoneMatch: ctx.IfNoneMatch,
				ETag:        etag,
			}, "work item type conditional query")
			if ctx.IfNoneMatch != nil && *ctx.IfNoneMatch == etag {
				return ctx.NotModified()
			}
		}
		ctx.ResponseData.Header().Set(LastModified, updatedAt.String())
		ctx.ResponseData.Header().Set(ETag, etag)
		ctx.ResponseData.Header().Set(CacheControl, MaxAge+"="+c.configuration.GetWorkItemTypeCacheControlMaxAge())
		// return the work item type along with conditional query and caching headers
		return ctx.OK(res)
	})
}

// Create runs the create action.
func (c *WorkitemtypeController) Create(ctx *app.CreateWorkitemtypeContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		var fields = map[string]app.FieldDefinition{}
		for key, fd := range ctx.Payload.Data.Attributes.Fields {
			fields[key] = *fd
		}
		wit, err := appl.WorkItemTypes().Create(ctx.Context, *ctx.Payload.Data.Relationships.Space.Data.ID, ctx.Payload.Data.ID, ctx.Payload.Data.Attributes.ExtendedTypeName, ctx.Payload.Data.Attributes.Name, ctx.Payload.Data.Attributes.Description, ctx.Payload.Data.Attributes.Icon, fields)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		ctx.ResponseData.Header().Set("Location", app.WorkitemtypeHref(wit.Data.ID))
		return ctx.Created(wit)
	})
}

// List runs the list action
func (c *WorkitemtypeController) List(ctx *app.ListWorkitemtypeContext) error {
	start, limit, err := parseLimit(ctx.Page)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errs.Wrap(err, "Could not parse paging"))
	}
	return application.Transactional(c.db, func(appl application.Application) error {
		result, err := appl.WorkItemTypes().List(ctx.Context, start, &limit)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, errs.Wrap(err, "Error listing work item types"))
		}
		return ctx.OK(result)
	})
}

// ListSourceLinkTypes runs the list-source-link-types action.
func (c *WorkitemtypeController) ListSourceLinkTypes(ctx *app.ListSourceLinkTypesWorkitemtypeContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		// Test that work item type exists
		_, err := appl.WorkItemTypes().Load(ctx.Context, ctx.WitID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		// Fetch all link types where this work item type can be used in the
		// source of the link
		res, err := appl.WorkItemLinkTypes().ListSourceLinkTypes(ctx.Context, ctx.WitID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		// Enrich link types
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkTypeHref, nil)
		err = enrichLinkTypeList(linkCtx, res)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		return ctx.OK(res)
	})
}

// ListTargetLinkTypes runs the list-target-link-types action.
func (c *WorkitemtypeController) ListTargetLinkTypes(ctx *app.ListTargetLinkTypesWorkitemtypeContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		// Test that work item type exists
		_, err := appl.WorkItemTypes().Load(ctx.Context, ctx.WitID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		// Fetch all link types where this work item type can be used in the
		// target of the linkg
		res, err := appl.WorkItemLinkTypes().ListTargetLinkTypes(ctx.Context, ctx.WitID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		// Enrich link types
		linkCtx := newWorkItemLinkContext(ctx.Context, appl, c.db, ctx.RequestData, ctx.ResponseData, app.WorkItemLinkTypeHref, nil)
		err = enrichLinkTypeList(linkCtx, res)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		return ctx.OK(res)
	})
}
