package resource

import (
	"context"

	"github.com/shashimalcse/cronuseo/internal/mongo_entity"
	"github.com/shashimalcse/cronuseo/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Service interface {
	Get(ctx context.Context, org_id string, id string) (Resource, error)
	Query(ctx context.Context, org_id string, filter Filter) ([]Resource, error)
	QueryActions(ctx context.Context, org_id string, filter Filter) ([]Action, error)
	Create(ctx context.Context, org_id string, input CreateResourceRequest) (Resource, error)
	Update(ctx context.Context, org_id string, id string, input UpdateResourceRequest) (Resource, error)
	Patch(ctx context.Context, org_id string, id string, input PatchResourceRequest) (Resource, error)
	Delete(ctx context.Context, org_id string, id string) error
}

type Resource struct {
	mongo_entity.Resource
}

type CreateResourceRequest struct {
	Identifier  string                    `json:"identifier" bson:"identifier"`
	DisplayName string                    `json:"display_name" bson:"display_name"`
	Actions     []mongo_entity.Action     `json:"actions,omitempty" bson:"actions"`
	Type        mongo_entity.ResourceType `json:"type,omitempty" bson:"type"`
}

func (m CreateResourceRequest) Validate() error {

	return validation.ValidateStruct(&m,
		validation.Field(&m.Identifier, validation.Required),
	)
}

type UpdateResourceRequest struct {
	DisplayName *string `json:"display_name" bson:"display_name"`
}

type PatchResourceRequest struct {
	AddedActions   []mongo_entity.Action `json:"added_actions,omitempty" bson:"added_actions"`
	RemovedActions []string              `json:"removed_actions,omitempty" bson:"removed_actions"`
}

type UpdateResource struct {
	DisplayName *string `json:"display_name" bson:"display_name"`
}

type PatchResource struct {
	AddedActions   []mongo_entity.Action `json:"added_actions,omitempty" bson:"added_actions"`
	RemovedActions []string              `json:"removed_actions,omitempty" bson:"removed_actions"`
}

type Action struct {
	Action   string `json:"action" bson:"action"`
	Resource string `json:"resource" bson:"resource"`
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {

	return service{repo: repo, logger: logger}
}

// Get resource by id.
func (s service) Get(ctx context.Context, org_id string, id string) (Resource, error) {

	resource, err := s.repo.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Error("Error while getting the resource.", zap.String("organization_id", org_id), zap.String("resource_id", id))
		return Resource{}, &util.NotFoundError{Path: "Resource"}
	}
	return Resource{*resource}, err
}

// Create new resource.
func (s service) Create(ctx context.Context, org_id string, req CreateResourceRequest) (Resource, error) {

	// Validate resource request.
	if err := req.Validate(); err != nil {
		s.logger.Error("Error while validating resource create request.")
		return Resource{}, &util.InvalidInputError{Path: "Invalid input for resource."}
	}

	// Check resource already exists.
	exists, _ := s.repo.CheckResourceExistsByIdentifier(ctx, org_id, req.Identifier)
	if exists {
		s.logger.Debug("Resource already exists.")
		return Resource{}, &util.AlreadyExistsError{Path: "Resource : " + req.Identifier}
	}
	resId := primitive.NewObjectID()
	actions := []mongo_entity.Action{}
	for _, action := range req.Actions {
		actionId := primitive.NewObjectID()
		actions = append(actions, mongo_entity.Action{
			ID:          actionId,
			Identifier:  action.Identifier,
			DisplayName: action.DisplayName,
		})
	}
	err := s.repo.Create(ctx, org_id, mongo_entity.Resource{
		ID:          resId,
		Identifier:  req.Identifier,
		DisplayName: req.DisplayName,
		Actions:     actions,
		Type:        req.Type,
	})
	if err != nil {
		s.logger.Info(err.Error())
		s.logger.Error("Error while creating resource.", zap.String("organization_id", org_id), zap.String("resource identifier", req.Identifier))
		return Resource{}, err
	}
	return s.Get(ctx, org_id, resId.Hex())
}

// Update resource.
func (s service) Update(ctx context.Context, org_id string, id string, req UpdateResourceRequest) (Resource, error) {

	// Get resource to check resource exists.
	_, err := s.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Debug("Resource not exists.", zap.String("resource_id", id))
		return Resource{}, &util.NotFoundError{Path: "Resource " + id + " not exists."}
	}

	if err := s.repo.Update(ctx, org_id, id, UpdateResource{
		DisplayName: req.DisplayName,
	}); err != nil {
		s.logger.Error("Error while updating resource.",
			zap.String("organization_id", org_id),
			zap.String("resource_id", id))
		return Resource{}, err
	}
	updatedResource, err := s.repo.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Debug("Resource not exists.", zap.String("resource_id", id))
		return Resource{}, &util.NotFoundError{Path: "Resource " + id + " not exists."}
	}
	return Resource{*updatedResource}, nil
}

// Patch resource.
func (s service) Patch(ctx context.Context, org_id string, id string, req PatchResourceRequest) (Resource, error) {

	// Get resource.
	_, err := s.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Debug("Resource not exists.", zap.String("resource_id", id))
		return Resource{}, &util.NotFoundError{Path: "Resource " + id + " not exists."}
	}

	var addedActions []mongo_entity.Action
	for _, action := range req.AddedActions {
		already_added, _ := s.repo.CheckActionAlreadyAddedToResourceByIdentifier(ctx, org_id, id, action.Identifier)
		if !already_added {
			actionId := primitive.NewObjectID()
			addedActions = append(addedActions, mongo_entity.Action{
				ID:          actionId,
				Identifier:  action.Identifier,
				DisplayName: action.DisplayName,
			})
		} else {
			return Resource{}, &util.AlreadyExistsError{Path: "Action : " + action.Identifier + " already added to resource."}
		}
	}
	// Set removed actions ids.
	var removedActions []string
	for _, action := range req.RemovedActions {
		already_added, _ := s.repo.CheckActionExistsByIdentifier(ctx, org_id, id, action)
		if already_added {
			removedActions = append(removedActions, action)
		} else {
			return Resource{}, &util.NotFoundError{Path: "Action " + action + " not exists."}
		}
	}

	if err := s.repo.Patch(ctx, org_id, id, PatchResource{
		AddedActions:   addedActions,
		RemovedActions: removedActions,
	}); err != nil {
		s.logger.Error("Error while updating resource.",
			zap.String("organization_id", org_id),
			zap.String("resource_id", id))
		return Resource{}, err
	}
	updatedResource, err := s.repo.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Debug("Resource not exists.", zap.String("resource_id", id))
		return Resource{}, &util.NotFoundError{Path: "Resource " + id + " not exists."}
	}
	return Resource{*updatedResource}, nil
}

// Delete resource.
func (s service) Delete(ctx context.Context, org_id string, id string) error {

	_, err := s.Get(ctx, org_id, id)
	if err != nil {
		s.logger.Error("Resource not exists.", zap.String("resource_id", id))
		return &util.NotFoundError{Path: "Resource " + id + " not exists."}
	}
	if err = s.repo.Delete(ctx, org_id, id); err != nil {
		s.logger.Error("Error while deleting resource.",
			zap.String("organization_id", org_id),
			zap.String("resource_id", id))
		return err
	}
	return nil
}

// Pagination filter.
type Filter struct {
	Cursor int    `json:"cursor" query:"cursor"`
	Limit  int    `json:"limit" query:"limit"`
	Name   string `json:"name" query:"name"`
}

// Get all resources.
func (s service) Query(ctx context.Context, org_id string, filter Filter) ([]Resource, error) {

	result := []Resource{}
	items, err := s.repo.Query(ctx, org_id)
	if err != nil {
		s.logger.Error("Error while retrieving all resources.",
			zap.String("organization_id", org_id))
		return []Resource{}, err
	}

	for _, item := range *items {
		result = append(result, Resource{item})
	}
	return result, err
}

func (s service) QueryActions(ctx context.Context, org_id string, filter Filter) ([]Action, error) {

	actions := []Action{}
	resources, err := s.repo.QueryWithActions(ctx, org_id)
	if err != nil {
		s.logger.Error("Error while retrieving all resources.",
			zap.String("organization_id", org_id))
		return []Action{}, err
	}

	for _, resource := range *resources {
		for _, action := range resource.Actions {
			actions = append(actions, Action{Resource: resource.Identifier, Action: action.Identifier})
		}
	}
	return actions, err
}
