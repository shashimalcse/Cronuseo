package resource

import (
	"context"

	db "github.com/shashimalcse/cronuseo/internal/db/mongo"
	"github.com/shashimalcse/cronuseo/internal/mongo_entity"
	"github.com/shashimalcse/cronuseo/internal/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	Get(ctx context.Context, org_id string, id string) (*mongo_entity.Resource, error)
	Query(ctx context.Context, org_id string) (*[]mongo_entity.Resource, error)
	QueryWithActions(ctx context.Context, org_id string) (*[]mongo_entity.Resource, error)
	Create(ctx context.Context, org_id string, resource mongo_entity.Resource) error
	Update(ctx context.Context, org_id string, id string, update_resource UpdateResource) error
	Patch(ctx context.Context, org_id string, id string, patch_resource PatchResource) error
	Delete(ctx context.Context, org_id string, id string) error
	CheckResourceExistById(ctx context.Context, org_id string, id string) (bool, error)
	CheckResourceExistsByIdentifier(ctx context.Context, org_id string, key string) (bool, error)
	CheckActionAlreadyAddedToResourceByIdentifier(ctx context.Context, org_id string, resource_id string, action_identifier string) (bool, error)
	CheckActionExistsByIdentifier(ctx context.Context, org_id string, resource_identifier string, action_identifier string) (bool, error)
}

type repository struct {
	mongoClient *mongo.Client
	mongoColl   *mongo.Collection
}

func NewRepository(mongodb *db.MongoDB) Repository {

	orgCollection := mongodb.MongoClient.Database(mongodb.MongoConfig.DBName).Collection(mongodb.MongoConfig.OrganizationCollectionName)

	return repository{mongoClient: mongodb.MongoClient, mongoColl: orgCollection}
}

// Get resource by id.
func (r repository) Get(ctx context.Context, org_id string, id string) (*mongo_entity.Resource, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return nil, err
	}

	resId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Define filter to find the resource by its ID
	filter := bson.M{"_id": orgId, "resources._id": resId}
	projection := bson.M{"resources.$": 1}
	// Find the role document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
	if err := result.Err(); err != nil {
		return nil, err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &util.NotFoundError{Path: "Resource"}
		}
		return nil, err
	}

	return &org.Resources[0], nil
}

// Create new resource.
func (r repository) Create(ctx context.Context, org_id string, resource mongo_entity.Resource) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}
	// Update the APIResources array for the given organization
	filter := bson.M{"_id": orgId}
	update := bson.M{"$push": bson.M{"resources": resource}}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil

}

func (r repository) Update(ctx context.Context, org_id string, id string, update_resource UpdateResource) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	resId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	if update_resource.DisplayName != nil && *update_resource.DisplayName != "" {

		filter := bson.M{"_id": orgId, "resources._id": resId}
		update := bson.M{"$set": bson.M{"resources.$.display_name": *update_resource.DisplayName}}
		_, err := r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repository) Patch(ctx context.Context, org_id string, id string, patch_resource PatchResource) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	resId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// add actions
	if len(patch_resource.AddedActions) > 0 {

		filter := bson.M{"_id": orgId, "resources._id": resId}
		update := bson.M{"$push": bson.M{"resources.$.actions": bson.M{
			"$each": patch_resource.AddedActions,
		}}}
		_, err = r.mongoColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}

	if len(patch_resource.RemovedActions) > 0 {

		filter := bson.M{"_id": orgId, "resources._id": resId}
		update := bson.M{"$pull": bson.M{"resources.$.actions": bson.M{
			"identifier": bson.M{"$in": patch_resource.RemovedActions},
		}}}
		_, err := r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
		if err != nil {
			return err
		}
	}

	return nil
}

// Get all resources.
func (r repository) Query(ctx context.Context, org_id string) (*[]mongo_entity.Resource, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return nil, err
	}

	// Define filter to find the resource by its ID
	filter := bson.M{"_id": orgId}
	projection := bson.M{"resources.actions": 0}
	// Find the resource document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
	if err := result.Err(); err != nil {
		return nil, err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &util.NotFoundError{Path: "Resource"}
		}
		return nil, err
	}

	return &org.Resources, nil
}

func (r repository) QueryWithActions(ctx context.Context, org_id string) (*[]mongo_entity.Resource, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return nil, err
	}

	// Define filter to find the resource by its ID
	filter := bson.M{"_id": orgId}
	// Find the resource document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter)
	if err := result.Err(); err != nil {
		return nil, err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &util.NotFoundError{Path: "Resource"}
		}
		return nil, err
	}

	return &org.Resources, nil
}

// Delete existing resource.
func (r repository) Delete(ctx context.Context, org_id string, id string) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	resId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define filter to find the resource by its ID
	filter := bson.M{"_id": orgId}
	update := bson.M{"$pull": bson.M{"resources": bson.M{"_id": resId}}}
	// Find the resource document in the "organizations" collection
	result, err := r.mongoColl.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}

	// Check if the update operation modified any documents
	if result.ModifiedCount == 0 {
		return err
	}

	return nil
}

// Check if resource exists by id.
func (r repository) CheckResourceExistById(ctx context.Context, org_id string, id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	resId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "resources._id": resId}

	// Search for the resource in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter)

	// Check if the resource was found
	if result.Err() == nil {
		return true, nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else {
		return false, result.Err()
	}
}

// Check if resource exists by key.
func (r repository) CheckResourceExistsByIdentifier(ctx context.Context, org_id string, identifier string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}
	filter := bson.M{"_id": orgId, "resources.identifier": identifier}

	// Search for the resource in the "organizations" collection
	count, err := r.mongoColl.CountDocuments(context.Background(), filter)

	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// check user already added to role
func (r repository) CheckActionAlreadyAddedToResourceByIdentifier(ctx context.Context, org_id string, resource_id string, action_identifier string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	resourceId, err := primitive.ObjectIDFromHex(resource_id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "resources._id": resourceId}
	projection := bson.M{"resources.$": 1}
	org := mongo_entity.Organization{}
	// Search for the resource in the "organizations" collection
	err = r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).Decode(&org)
	if err != nil {
		return false, err
	}
	resource := org.Resources[0]
	// Check if the action identifier exists in the resource's Actions field
	for _, action := range resource.Actions {
		if action.Identifier == action_identifier {
			return true, nil
		}
	}

	// User ID not found in the resource's Actions field
	return false, nil
}

func (r repository) CheckActionExistsByIdentifier(ctx context.Context, org_id string, resource_id string, action_identifier string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	resourceId, err := primitive.ObjectIDFromHex(resource_id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "resources._id": resourceId, "resources.actions.identifier": action_identifier}

	result := r.mongoColl.FindOne(context.Background(), filter)

	// Check if the resource was found
	if result.Err() == nil {
		return true, nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else {
		return false, result.Err()
	}
}
