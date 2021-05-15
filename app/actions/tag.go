package actions

import (
	"context"
	"regexp"

	"github.com/getfider/fider/app/models/query"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/validate"
	"github.com/gosimple/slug"
)

var colorRegex = regexp.MustCompile(`^([A-Fa-f0-9]{6})$`)

// CreateEditTag is used to create a new tag or edit existing
type CreateEditTag struct {
	Tag   *models.Tag
	Model *models.CreateEditTag
}

// Returns the struct to bind the request to
func (action *CreateEditTag) BindTarget() interface{} {
	action.Model = new(models.CreateEditTag)
	return action.Model
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *CreateEditTag) IsAuthorized(ctx context.Context, user *models.User) bool {
	return user != nil && user.IsAdministrator()
}

// Validate if current model is valid
func (action *CreateEditTag) Validate(ctx context.Context, user *models.User) *validate.Result {
	result := validate.Success()

	if action.Model.Slug != "" {
		getSlug := &query.GetTagBySlug{Slug: action.Model.Slug}
		err := bus.Dispatch(ctx, getSlug)
		if err != nil {
			return validate.Error(err)
		}
		action.Tag = getSlug.Result
	}

	if action.Model.Name == "" {
		result.AddFieldFailure("name", "Name is required.")
	} else if len(action.Model.Name) > 30 {
		result.AddFieldFailure("name", "Name must have less than 30 characters.")
	} else {
		getDuplicateSlug := &query.GetTagBySlug{Slug: slug.Make(action.Model.Name)}
		err := bus.Dispatch(ctx, getDuplicateSlug)
		if err != nil && errors.Cause(err) != app.ErrNotFound {
			return validate.Error(err)
		} else if err == nil && (action.Tag == nil || action.Tag.ID != getDuplicateSlug.Result.ID) {
			result.AddFieldFailure("name", "This tag name is already in use.")
		}
	}

	if action.Model.Color == "" {
		result.AddFieldFailure("color", "Color is required.")
	} else if len(action.Model.Color) != 6 {
		result.AddFieldFailure("color", "Color must be exactly 6 characters.")
	} else if !colorRegex.MatchString(action.Model.Color) {
		result.AddFieldFailure("color", "Color is invalid.")
	}

	return result
}

// DeleteTag is used to delete an existing tag
type DeleteTag struct {
	Tag   *models.Tag
	Model *models.DeleteTag
}

// Returns the struct to bind the request to
func (action *DeleteTag) BindTarget() interface{} {
	action.Model = new(models.DeleteTag)
	return action.Model
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *DeleteTag) IsAuthorized(ctx context.Context, user *models.User) bool {
	return user != nil && user.IsAdministrator()
}

// Validate if current model is valid
func (action *DeleteTag) Validate(ctx context.Context, user *models.User) *validate.Result {
	getSlug := &query.GetTagBySlug{Slug: action.Model.Slug}
	err := bus.Dispatch(ctx, getSlug)
	if err != nil {
		return validate.Error(err)
	}

	action.Tag = getSlug.Result
	return validate.Success()
}

// AssignUnassignTag is used to assign or remove a tag to/from an post
type AssignUnassignTag struct {
	Tag   *models.Tag
	Post  *models.Post
	Model *models.AssignUnassignTag
}

// Returns the struct to bind the request to
func (action *AssignUnassignTag) BindTarget() interface{} {
	action.Model = new(models.AssignUnassignTag)
	return action.Model
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *AssignUnassignTag) IsAuthorized(ctx context.Context, user *models.User) bool {
	return user != nil && user.IsCollaborator()
}

// Validate if current model is valid
func (action *AssignUnassignTag) Validate(ctx context.Context, user *models.User) *validate.Result {
	getPost := &query.GetPostByNumber{Number: action.Model.Number}
	getSlug := &query.GetTagBySlug{Slug: action.Model.Slug}
	if err := bus.Dispatch(ctx, getPost, getSlug); err != nil {
		return validate.Error(err)
	}

	action.Post = getPost.Result
	action.Tag = getSlug.Result
	return validate.Success()
}
