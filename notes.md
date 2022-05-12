## Before Suite
- Build controller image
- Load image onto kind cluster (if running on KinD)

## Local tests
All local tests seem to consist of the same thing:
1. Run `make install`
2. Run `make run`
3. Kill the process running `make run`
4. Run `make uninstall`

Can this be simplified into a basic e2e test helper function?

## Cluster tests
Common operations when testing on the cluster:
Before each:
- Deploy with `make deploy`

After each:
- delete curl pod
- delete test CR instances
- cleanup permissions
- undeploy with `make undeploy`
- delete namespace

Tests:
- Verify that the controller is up and running
- Ensure ServiceMonitor is created
- Ensure metrics service is created
- create custom resource
- Grant permission to access metrics
- read metrics token
- create curl pod
- validate curl pod running
- check metrics is serving
- check that cr is reconciled

## Takeaways
There should be some generalized e2e helper functions that can be used to make the creation of e2e tests easier.

Some thoughts on actions that can be turned into helper functions:
- Deploying operator
- Clean up
- Getting Metrics
- Starting Prometheus Operator
- Build image
- Load image to KinD cluster