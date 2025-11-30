# Deploying local docker cluster

## Prerequisites
* [Docker Desktop](https://www.docker.com/products/docker-desktop/) - Docker 17.12.0+
* [Docker compose 2+](https://github.com/docker/compose/releases/tag/v2.14.1)

### `favobft` consensus
When deploying with `favobft` consensus, there are some additional dependencies:
* [npm](https://nodejs.org/en/)
* [go 1.18.x](https://go.dev/dl/)

## Local development
Running `favo-chain` local cluster with docker can be done very easily by using provided `scripts` folder
or by running `docker-compose` manually.

### Using provided `scripts` folder
***All commands need to be run from the repo root / root folder.***

* `scripts/cluster ibft --docker` - deploy environment with `ibft` consensus
* `scripts/cluster favobft --docker` - deploy environment with `favobft` consensus
* `scripts/cluster {ibft or favobft} --docker stop` - stop containers
* `scripts/cluster {ibft or favobft}--docker destroy` - destroy environment (delete containers and volumes)

### Using `docker-compose`
***All commands need to be run from the repo root / root folder.***

#### use `ibft` PoA consensus
* `export EDGE_CONSENSUS=ibft` - set `ibft` consensus
* `docker-compose -f ./docker/local/docker-compose.yml up -d --build` - deploy environment

#### use `favobft` consensus
* `cd core-contracts && npm install && npm run compile && cd -` - install `npm` dependencies and compile smart contracts
* `go run ./consensus/favobft/contractsapi/artifacts-gen/main.go` generate needed code
* `export EDGE_CONSENSUS=favobft` - set `favobft` consensus
* `docker-compose -f ./docker/local/docker-compose.yml up -d --build` - deploy environment

#### stop / destroy 
* `docker-compose -f ./docker/local/docker-compose.yml stop` - stop containers
* `docker-compose -f ./docker/local/docker-compose.yml down -v` - destroy environment

## Customization
Use `docker/local/favo-chain.sh` script to customize chain parameters.    
All parameters can be defined at the very beginning of the script, in the `CHAIN_CUSTOM_OPTIONS` variable.   
It already has some default parameters, which can be easily modified. 
These are the `genesis` parameters from the official [docs](https://wiki.favo.technology/docs/edge/get-started/cli-commands#genesis-flags).  

Primarily, the `--premine` parameter needs to be edited to include the accounts that the user has access to.   

## Considerations

### Submodules
Before deploying `favobft` environment, `core-contracts` submodule needs to be downloaded.  
To do that simply run `make download-submodules`.

### Build times
When building containers for the first time (or after purging docker build cache)
it might take a while to complete, depending on the hardware that the build operation is running on.

### Production
This is **NOT** a production ready deployment. It is to be used in *development* / *test* environments only.       
For production usage, please check out the official [docs](https://wiki.favo.technology/docs/edge/overview/). 
