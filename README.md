<p align="center">
<img alt="Screenshot 2022-11-05 at 14 41 11" src="https://user-images.githubusercontent.com/43197743/205511091-bed0ace7-705d-4372-8980-872bcd71a200.png">
</p>

<h1 align="center">Authorization as a Service</h1>
<h2 align="center">Let's make auth-z easy!</h2>

## What is cronuseo ?

Cronuseo is an open-source authorization solution that allows developers to easily integrate permissions and access control into their products within minutes.

> Example: A developer can call the cronuseo and get a clear answer if User A has the permissions to create Resource B.

cronuseo is based on modern, open-source foundation which includes Open Policy Agent (OPA), Zanzibar.

## Main features:

* Role-based Access Control (RBAC)

## Get started

> Note : cronuseo still in the experimental stage. Only tested in the local environemnt.

### Set up the local environment

* ``` curl -LJO https://raw.githubusercontent.com/shashimalcse/cronuseo/HEAD/docker-compose-db.yml | curl -LJO https://raw.githubusercontent.com/shashimalcse/cronuseo/HEAD/docker-compose.yml ```
* Prepare a [mongodb](https://hub.docker.com/_/mongo) instance ``` docker compose -f docker-compose-database.yml up```
* Make sure to update the necessary configuration in the `config/local.yml` file, and don't forget to replace the jwks endpoint with the ones provided by your own identity provider. (only tested with [asgardeo](https://wso2.com/asgardeo/))
* Start management server and check server (Policy Decision Point) ``` docker compose up --build```

## cronuseo SDKs for applications
use these sdks to check permissions for the user.
* python - https://pypi.org/project/cronuseosdk
* nodejs - https://www.npmjs.com/package/cronuseosdk
* golang - https://github.com/shashimalcse/cronuseogosdk

