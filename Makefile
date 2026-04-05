.PHONY: generate

generate:
	oapi-codegen --config oapi-codegen.yaml api.yaml
