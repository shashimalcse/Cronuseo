package user

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
	Get(ctx context.Context, org_id string, id string) (*mongo_entity.User, error)
	GetIdByIdentifier(ctx context.Context, org_id string, identifier string) (string, error)
	Query(ctx context.Context, org_id string) (*[]mongo_entity.User, error)
	Create(ctx context.Context, org_id string, user mongo_entity.User) error
	Update(ctx context.Context, org_id string, id string, update_user UpdateUser) error
	Patch(ctx context.Context, org_id string, id string, req PatchUser) error
	Delete(ctx context.Context, org_id string, id string) error
	CheckUserExistById(ctx context.Context, org_id string, id string) (bool, error)
	CheckUserExistsByIdentifier(ctx context.Context, org_id string, key string) (bool, error)
	CheckRoleExistById(ctx context.Context, org_id string, id string) (bool, error)
	CheckRoleAlreadyAssignToUserById(ctx context.Context, org_id string, user_id string, role_id string) (bool, error)
	CheckGroupExistById(ctx context.Context, org_id string, id string) (bool, error)
	CheckGroupAlreadyAssignToUserById(ctx context.Context, org_id string, user_id string, group_id string) (bool, error)
}

type repository struct {
	mongoClient *mongo.Client
	mongoColl   *mongo.Collection
}

func NewRepository(mongodb *db.MongoDB) Repository {

	orgCollection := mongodb.MongoClient.Database(mongodb.MongoConfig.DBName).Collection(mongodb.MongoConfig.OrganizationCollectionName)

	return repository{mongoClient: mongodb.MongoClient, mongoColl: orgCollection}
}

// Get user by id.
func (r repository) Get(ctx context.Context, org_id string, id string) (*mongo_entity.User, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return nil, err
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Define filter to find the user by its ID
	filter := bson.M{"_id": orgId, "users._id": userId}
	projection := bson.M{"users.$": 1}
	// Find the user document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
	if err := result.Err(); err != nil {
		return nil, err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &util.NotFoundError{Path: "User"}
		}
		return nil, err
	}

	return &org.Users[0], nil
}

// Get user id by identifier.
func (r repository) GetIdByIdentifier(ctx context.Context, org_id string, identifier string) (string, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return "", err
	}

	// Define filter to find the user by its ID
	filter := bson.M{"_id": orgId, "users.identifier": identifier}
	projection := bson.M{"users.$": 1}
	// Find the user document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
	if err := result.Err(); err != nil {
		return "", err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", &util.NotFoundError{Path: "User"}
		}
		return "", err
	}

	return org.Users[0].ID.Hex(), nil
}

// Create new user.
func (r repository) Create(ctx context.Context, org_id string, user mongo_entity.User) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}
	// Update the APIResources array for the given organization
	filter := bson.M{"_id": orgId}
	update := bson.M{"$push": bson.M{"users": user}}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil

}

func (r repository) Update(ctx context.Context, org_id string, id string, update_user UpdateUser) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": orgId, "users._id": userId}
	update := bson.M{"$set": bson.M{}}
	if update_user.UserProperties != nil {
		update["$set"].(bson.M)["users.$.user_properties"] = *&update_user.UserProperties
	}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (r repository) Patch(ctx context.Context, org_id string, id string, patch_user PatchUser) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": orgId, "users._id": userId}
	updates := bson.M{}
	for key, value := range patch_user.UserProperties {
		updates["users.$.user_properties."+key] = value
	}

	update := bson.M{"$set": updates}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// add roles
	if len(patch_user.AddedRoles) > 0 {

		filter := bson.M{"_id": orgId, "users._id": userId}
		update := bson.M{"$push": bson.M{"users.$.roles": bson.M{
			"$each": patch_user.AddedRoles,
		}}}
		_, err = r.mongoColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}

		for _, roleId := range patch_user.AddedRoles {
			filter := bson.M{"_id": orgId, "roles._id": roleId}
			update := bson.M{"$addToSet": bson.M{"roles.$.users": userId}}
			_, err = r.mongoColl.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}

	}

	// remove roles
	if len(patch_user.RemovedRoles) > 0 {

		filter := bson.M{"_id": orgId, "users._id": userId}
		update := bson.M{"$pull": bson.M{"users.$.roles": bson.M{"$in": patch_user.RemovedRoles}}}
		_, err := r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
		if err != nil {
			return err
		}

		for _, roleId := range patch_user.RemovedRoles {
			filter := bson.M{"_id": orgId, "roles._id": roleId}
			update := bson.M{"$pull": bson.M{"roles.$.users": userId}}
			_, err = r.mongoColl.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}
	}

	// add groups
	if len(patch_user.AddedGroups) > 0 {

		filter := bson.M{"_id": orgId, "users._id": userId}
		update := bson.M{"$push": bson.M{"users.$.groups": bson.M{
			"$each": patch_user.AddedGroups,
		}}}
		_, err = r.mongoColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}

		for _, groupId := range patch_user.AddedGroups {
			filter := bson.M{"_id": orgId, "groups._id": groupId}
			update := bson.M{"$addToSet": bson.M{"groups.$.users": userId}}
			_, err = r.mongoColl.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}

	}

	// remove groups
	if len(patch_user.RemovedGroups) > 0 {

		filter := bson.M{"_id": orgId, "users._id": userId}
		update := bson.M{"$pull": bson.M{"users.$.groups": bson.M{"$in": patch_user.RemovedGroups}}}
		_, err := r.mongoColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
		if err != nil {
			return err
		}

		for _, groupId := range patch_user.RemovedGroups {
			filter := bson.M{"_id": orgId, "groups._id": groupId}
			update := bson.M{"$pull": bson.M{"groups.$.users": userId}}
			_, err = r.mongoColl.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete existing user.
func (r repository) Delete(ctx context.Context, org_id string, id string) error {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return err
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define filter to find the user by its ID
	filter := bson.M{"_id": orgId}
	update := bson.M{"$pull": bson.M{"users": bson.M{"_id": userId}}}
	// Find the user document in the "organizations" collection
	result, err := r.mongoColl.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}

	// Check if the update operation modified any documents
	if result.ModifiedCount == 0 {
		return err
	}

	filter = bson.M{"_id": orgId}
	update = bson.M{"$pull": bson.M{"groups.$[].users": userId}}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	filter = bson.M{"_id": orgId}
	update = bson.M{"$pull": bson.M{"roles.$[].users": userId}}
	_, err = r.mongoColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Get all users.
func (r repository) Query(ctx context.Context, org_id string) (*[]mongo_entity.User, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return nil, err
	}

	// Define filter to find the user by its ID
	filter := bson.M{"_id": orgId}
	projection := bson.M{"users.roles": 0, "users.groups": 0}
	// Find the user document in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
	if err := result.Err(); err != nil {
		return nil, err
	}

	// Decode the organization document into a struct
	var org mongo_entity.Organization
	if err := result.Decode(&org); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &util.NotFoundError{Path: "User"}
		}
		return nil, err
	}

	return &org.Users, nil
}

// Check if user exists by id.
func (r repository) CheckUserExistById(ctx context.Context, org_id string, id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "users._id": userId}

	// Search for the user in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter)

	// Check if the user was found
	if result.Err() == nil {
		return true, nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else {
		return false, result.Err()
	}
}

// Check if user exists by key.
func (r repository) CheckUserExistsByIdentifier(ctx context.Context, org_id string, identifier string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}
	filter := bson.M{"_id": orgId, "users.username": identifier}

	// Search for the user in the "organizations" collection
	count, err := r.mongoColl.CountDocuments(context.Background(), filter)

	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// Check if role exists by id.
func (r repository) CheckRoleExistById(ctx context.Context, org_id string, id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	roleId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "roles._id": roleId}

	// Search for the role in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter)

	// Check if the role was found
	if result.Err() == nil {
		return true, nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else {
		return false, result.Err()
	}
}

// Check if role already assign to user by id.
func (r repository) CheckRoleAlreadyAssignToUserById(ctx context.Context, org_id string, user_id string, role_id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	userId, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return false, err
	}

	roleId, err := primitive.ObjectIDFromHex(role_id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "users._id": userId}
	projection := bson.M{"users.$": 1}
	org := mongo_entity.Organization{}
	// Search for the role in the "organizations" collection
	err = r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).Decode(&org)
	if err != nil {
		return false, err
	}
	user := org.Users[0]
	// Check if the role ID exists in the user's Roles field
	for _, r := range user.Roles {
		if r == roleId {
			return true, nil
		}
	}

	// Role ID not found in the user's Roles field
	return false, nil
}

// Check if group exists by id.
func (r repository) CheckGroupExistById(ctx context.Context, org_id string, id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	groupId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "groups._id": groupId}

	// Search for the group in the "organizations" collection
	result := r.mongoColl.FindOne(context.Background(), filter)

	// Check if the group was found
	if result.Err() == nil {
		return true, nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else {
		return false, result.Err()
	}
}

// Check if group already assign to user by id.
func (r repository) CheckGroupAlreadyAssignToUserById(ctx context.Context, org_id string, user_id string, group_id string) (bool, error) {

	orgId, err := primitive.ObjectIDFromHex(org_id)
	if err != nil {
		return false, err
	}

	userId, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return false, err
	}

	groupId, err := primitive.ObjectIDFromHex(group_id)
	if err != nil {
		return false, err
	}

	filter := bson.M{"_id": orgId, "users._id": userId}
	projection := bson.M{"users.$": 1}
	org := mongo_entity.Organization{}
	// Search for the role in the "organizations" collection
	err = r.mongoColl.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).Decode(&org)
	if err != nil {
		return false, err
	}
	user := org.Users[0]
	// Check if the group ID exists in the user's Groups field
	for _, r := range user.Groups {
		if r == groupId {
			return true, nil
		}
	}

	// Group ID not found in the user's Groups field
	return false, nil
}
