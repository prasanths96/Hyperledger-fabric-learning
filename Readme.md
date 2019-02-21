# Config :
	- 4 Orgs : bda, fa, la, orderer
	- 1 peer each: peer0
	- orderer peer named orderer0
	- 3 CAs, one for each org: bda, fa, la
	- Setup container registers nodes to respective CAs
	- Cli container is used for playing with chaincode.

# Steps to run network:
	## Modifying docker-composer.yml
	- Open HLF-multi-layered-network/layout-approval-network/docker-compose.yml
	- Sroll to bottom:
		- Find in cli.volumes:
			- /home/osgdev/fabric-demos/HLF-multi-layered-network/chaincode
		- replace the above path with:
			- {Your project path where HLF-multi-layered-network folder is present}/HLF-multi-layered-network/chaincode

	- Save and close

	## Running the script:
	
	- open terminal inside: HLF-multi-layered-network/layout-approval-network
	- run "./start.sh"

	## CLI container:

	- Wait for 1-2 minutes for CAs to start and nodes to get registered with CAs
	- Open cli container: "docker exec -it cli bash"

	## Run the script to complete process until instantiating of chaincode:

	- run ". /scripts/cli-script.sh" inside the cli container. (dot)<space> is very important in this command as exports inside the script need to be in the same bash.

	## Now, follow the "HLF-multi-layered-network/layout-approval-network/scripts/Mannual-Commands.sh" file

	## To stop:
	
	- run "sudo ./stop.sh" (su rights needed to remove folders created by containers)
	- Remove chaincode images starting with "dev-" by : "docker rmi dev-..."








