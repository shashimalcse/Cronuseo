CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
SELECT uuid_generate_v4();
CREATE TABLE if not exists organization(
   org_id uuid DEFAULT uuid_generate_v4 (),
   org_key VARCHAR(40) NOT NULL,
   name VARCHAR(40) NOT NULL,
   PRIMARY KEY ( org_id ));
    
  CREATE TABLE if not exists project(
   project_id uuid DEFAULT uuid_generate_v4 (),
   project_key VARCHAR(40) NOT NULL,
   name VARCHAR(40),
   org_id uuid,
   PRIMARY KEY ( project_id ),
   CONSTRAINT FK_organization_project FOREIGN KEY(org_id) REFERENCES organization(org_id)
   ); 
  
 CREATE TABLE if not exists organization_user(
   user_id uuid DEFAULT uuid_generate_v4 (),
   username VARCHAR(40) NOT NULL,
   first_name VARCHAR(100),
   last_name VARCHAR(100),
   org_id uuid,
   PRIMARY KEY ( user_id ),
   CONSTRAINT FK_organization_org_user FOREIGN KEY(org_id) REFERENCES organization(org_id)
   );
 
 CREATE TABLE if not exists organization_group(
   group_id uuid DEFAULT uuid_generate_v4 (),
   group_key VARCHAR(40) NOT NULL,
   name VARCHAR(40) NOT NULL,
   org_id uuid,
   PRIMARY KEY ( group_id ),
   CONSTRAINT FK_organization_org_group FOREIGN KEY(org_id) REFERENCES organization(org_id)
   );
   
 CREATE TABLE if not exists group_user(
   group_id uuid,
   user_id uuid,
   CONSTRAINT FK_group_user_organization_user FOREIGN KEY(user_id) references organization_user(user_id),
   CONSTRAINT FK_group_user_organization_group FOREIGN KEY(group_id) REFERENCES organization_group(group_id)
   );
   
 CREATE TABLE if not exists resource (
   resource_id uuid DEFAULT uuid_generate_v4 (),
   resource_key VARCHAR(40) NOT NULL,
   name VARCHAR(40),
   project_id uuid,
   PRIMARY KEY ( resource_id ),
   CONSTRAINT FK_project_resource FOREIGN KEY(project_id) REFERENCES project(project_id)
   ); 
   
CREATE TABLE if not exists resource_action (
   resource_action_id uuid DEFAULT uuid_generate_v4 (),
   resource_action_key VARCHAR(40) NOT NULL,
   name VARCHAR(40),
   resource_id uuid,
   PRIMARY KEY ( resource_action_id ),
   CONSTRAINT FK_resource_resource_action FOREIGN KEY(resource_id) REFERENCES resource(resource_id)
   ); 
   
CREATE TABLE if not exists resource_role (
   resource_role_id uuid DEFAULT uuid_generate_v4 (),
   resource_role_key VARCHAR(40) NOT NULL,
   name VARCHAR(40),
   resource_id uuid,
   PRIMARY KEY ( resource_role_id ),
   CONSTRAINT FK_resource_resource_role FOREIGN KEY(resource_id) REFERENCES resource(resource_id)
   ); 
   
 CREATE TABLE if not exists action_role(
   resource_role_id uuid,
   resource_action_id uuid,
   resource_id uuid,
   CONSTRAINT FK_action_role_resource FOREIGN KEY(resource_id) REFERENCES resource(resource_id),
   CONSTRAINT FK_action_role_resource_role FOREIGN KEY(resource_role_id) REFERENCES resource_role(resource_role_id),
   CONSTRAINT FK_action_role_resource_action FOREIGN KEY(resource_action_id) REFERENCES resource_action(resource_action_id)
   );
   
CREATE TABLE if not exists user_resource_role(
   user_id uuid,
   resource_role_id uuid,
   CONSTRAINT FK_resource_role_user_resource_role FOREIGN KEY(resource_role_id) REFERENCES resource_role(resource_role_id),
   CONSTRAINT FK_organization_user_user_resource_role FOREIGN KEY(user_id) REFERENCES organization_user(user_id)
   );
   
CREATE TABLE if not exists group_resource_role(
   group_id uuid,
   resource_role_id uuid,
   CONSTRAINT FK_resource_role_group_resource_role FOREIGN KEY(resource_role_id) REFERENCES resource_role(resource_role_id),
   CONSTRAINT FK_organization_user_group_resource_role FOREIGN KEY(group_id) REFERENCES organization_group(group_id)
   );
   
CREATE TABLE if not exists resource_action_role(
   resource_action_id uuid,
   resource_role_id uuid,
   resource_id uuid,
   CONSTRAINT FK_resource_action_role_role FOREIGN KEY(resource_role_id) REFERENCES resource_role(resource_role_id),
   CONSTRAINT FK_resource_action_role_action FOREIGN KEY(resource_action_id) REFERENCES resource_action(resource_action_id),
   CONSTRAINT FK_resource_action_role_resource FOREIGN KEY(resource_id) REFERENCES resource(resource_id)
   );
  
CREATE TABLE if not exists resource_action_role_key(
   resource_action_key VARCHAR(40) NOT NULL,
   resource_role_key VARCHAR(40) NOT NULL,
   resource_key VARCHAR(160) NOT NULL
   );
   