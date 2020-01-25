# Config :
	- 4 Orgs : bda, fa, la, orderer
	- 1 peer each: peer0
	- orderer peer named orderer0
	- 3 CAs, one for each org: bda, fa, la
	- Setup container registers nodes to respective CAs
	- Cli container is used for playing with chaincode.

# Steps to run network:
	## Modifying docker-composer.yml
	- Open ./layout-approval-network/docker-compose.yml
	- Sroll to bottom:
		- Find in cli.volumes:
			- /home/osgdev/fabric-demos/HLF-multi-layered-network/chaincode
		- replace the above path with:
			- {Your project path where this repo folder is present}/chaincode

	- Save and close

	## Running the script:
	
	- open terminal inside: ./layout-approval-network
	- run "./start.sh"

	## CLI container:

	- Wait for 1-2 minutes for CAs to start and nodes to get registered with CAs
	- Open cli container: "docker exec -it cli bash"

	## Run the script to complete process until instantiating of chaincode:

	- run ". /scripts/cli-script.sh" inside the cli container. (dot)<space> is very important in this command as exports env variables inside the script need to be in the same bash. If not, usually a new bash instance will be created for each individual scripts that are inside the master script, thereby losing the exported env variables beforehand.

	## Now, follow the "./layout-approval-network/scripts/Mannual-Commands.sh" file

	## To stop:
	
	- run "sudo ./stop.sh" (su rights needed to remove folders created by containers)
	- Remove chaincode images starting with "dev-" by : "docker rmi dev-..."








