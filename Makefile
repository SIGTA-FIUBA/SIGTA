.PHONY: contracts contracts-backend contracts-web

# Regenerates the code of every side from contracts/openapi.yaml.
contracts: contracts-backend contracts-web

contracts-backend:
	cd services/backend/src/adapters/rest && go tool oapi-codegen -config oapi-codegen.yaml ../../../../../contracts/openapi.yaml

contracts-web:
	cd web && npm run generate
