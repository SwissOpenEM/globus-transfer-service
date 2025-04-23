# Globus Transfer Service

### Summary

This service allows for requesting [Globus](https://www.globus.org) transfers using the [Scicat](https://scicatproject.github.io) token. It also keeps track of ongoing transfers, and their state can be polled from this service.

It relies on a single [service account](https://docs.globus.org/guides/recipes/automate-with-service-account/) to request and track transfers, which can be set using environment variables.

It also assumes that there's a set of possible destinations and sources. Whether it's possible to ingest from the source to the given destination is defined by the Scicat user's current set of groups. The group associations are given to this service as a config. 

### Configuration

You can find an example of the settings at `[PROJECT_ROOT]/example-conf.yaml`

 - `scicatUrl` - the **base** url fo the instance of scicat to use (without the `/api/v[X]` part)
 - `facilityCollectionIDs` - a map of facility names (identifiers) to their collection id's
 - `globusScopes` - the scopes to use for the client connection. Access is required to transfer api and specific collections
 - `port` - the port at which the server should run
 - `facilitySrcGroupTemplate` - the template to use for groups (their names) that allow users to use facilities listed in `facilityCollectionIDs` as the source of their transfer requests
 - `facilityDstGroupTemplate` - same as above, but as the destination of their transfer requests
 - `destinationPathTemplate` - the template to use for determining the path at the destination of the transfer
 - `task` - a set of settings for configuring the handling of transfer tasks
   - `maxConcurrency` - maximum number of transfer tasks executed in parallel
   - `queueSize` - how many tasks can be put in a queue (0 is infinite) 
   - `pollInterval` - the amount of seconds to wait before a task polls Globus again to update the status of the transfer


### Environment variables

 - `GLOBUS_CLIENT_ID` - the client id for the service account (2-legged OAUTH, trusted client model)
 - `GLOBUS_CLIENT_SECRET` - the client secret for the service account (2-legged OAUTH, trusted client model)
 - `SCICAT_SERVICE_USER_USERNAME` - the username for the service user to use for creating transfer jobs in scicat
 - `SCICAT_SERVICE_USER_PASSWORD` - the above user's password