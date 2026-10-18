.PHONY: contracts

contracts: 
	cd services/backend/src/adapters/rest && go tool oapi-codegen -config oapi-codegen.yaml ../../../../../contracts/openapi.yaml
