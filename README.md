# Stock Picker
---

## Run locally
`bash
ALPHA_VANTAGE_URL=https://www.alphavantage.co NDAYS=5 API_KEY={REDACTED} SYMBOL=MSFT make run
`

## Run
`bash
make all
`

## Help
`bash
make help
`

## API Contracts
Check [swagger endpoint](http://localhost:8080/swagger/index.html) after running `make run`

## TODO
To increase the resilience of the application:
- [ ] Monitoring and alerting for metrics like:
        * compute resources
        * SLX for responses from vendor
        * SLX for responses from application
- [ ] Logging: Enable info and debug log and ideally have a reliable sink system to push the logging to (ex: Kabana/Splunk/etc.).
- [ ] Resiliance in CICD: Add CICD pipelines by adding following capabilities:
        * enable static analysis of code and container for vulnerability
        * enable security scans for kubernetes manifest
        * add unit testing as a stage gate for CI failures
        * add integration and smoke testing and append it to CI stages. integration testing enables merges to main branch and smoke testing should enable promotion to staging and production environments.
- [ ] Secret Management: Secrets should be injected via an init container (ex: vault) and not be hardcoded or exposed. No secrets should be set via environment variables
- [ ] Scaling: Setup autoscaling based on metrics like resource constraints, etc.
- [ ] Add HTTPS for Layer7 protection; although this may be abstracted away into ingress workflow if Service Mesh is enabled at a platform layer.

