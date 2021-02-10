# Angelie
## About
Angelie is a component of project Poseidon (a connected swimming pool system)  
It's responsible to decode telemetries received throught MQTT/HTTP message and forward them to [Athena](https://github.com/AlexisMtr/athena)

> NOTE: Poseidon is a training tool to form about DevOps practices (logging, monitoring, cloud, CI/CD, IaC, GitOps, docker, k8s ...)

## Documentation
You can find more documentation on subfolders README
* [Configure, Build and Run](./src/README.md)
* [Deploy on k8s](./helm/README.md)
* [Deploy on Azure](./infra/azure/README.md)

## Changelog
You can find `CHANGELOG.md` on each subfolder
* [Athena Changelog](./src/CHANGELOG.md)
* [Helm Chart Changelog](./helm/CHANGELOG.md)
* [Azure Terraform Changelog](./infra/azure/terraform/CHANGELOG.md)
* [Azure ARM Changelog](./infra/azure/arm/CHANGELOG.md)