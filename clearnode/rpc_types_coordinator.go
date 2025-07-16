package main

import (
	"fmt"
	"sort"
	"strings"
	
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SchemaOrchestrator struct {
	schemas              map[string]SchemaInfo
	allDefinitions       map[string]SchemaProperty
	commonDefinitions    map[string]SchemaProperty
	requestDefinitions   map[string]SchemaProperty
	responseDefinitions  map[string]SchemaProperty
	requestTypeMappings  map[string]RPCMethod // typeName -> rpcMethod
	responseTypeMappings map[string]RPCMethod // typeName -> rpcMethod
	codeFileGenerator    *CodeFileGenerator
}

func NewSchemaOrchestrator() *SchemaOrchestrator {
	return &SchemaOrchestrator{
		schemas:              make(map[string]SchemaInfo),
		allDefinitions:       make(map[string]SchemaProperty),
		commonDefinitions:    make(map[string]SchemaProperty),
		requestDefinitions:   make(map[string]SchemaProperty),
		responseDefinitions:  make(map[string]SchemaProperty),
		requestTypeMappings:  make(map[string]RPCMethod),
		responseTypeMappings: make(map[string]RPCMethod),
	}
}

func (orchestrator *SchemaOrchestrator) LoadSchemas(requestDirectoryPath, responseDirectoryPath string) error {
	loadedSchemas, allDefinitions, err := LoadSchemas(requestDirectoryPath, responseDirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	orchestrator.schemas = loadedSchemas
	orchestrator.allDefinitions = allDefinitions
	return nil
}

func (orchestrator *SchemaOrchestrator) CategorizeDefinitions() {
	requestUsageMap, responseUsageMap := orchestrator.createUsageTracker()
	orchestrator.categorizeDefinitionsByUsage(requestUsageMap, responseUsageMap)
	orchestrator.buildRPCMethodMappings()
}

func (orchestrator *SchemaOrchestrator) GenerateAllFiles(schemaDirectoryPath string, sdkRootDirectoryPath string) error {
	config, err := NewGenerationConfig(schemaDirectoryPath, sdkRootDirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to create generation config: %w", err)
	}

	dependencies := orchestrator.createGeneratorDependencies()
	codeFileGenerator, err := NewCodeFileGenerator(dependencies)
	if err != nil {
		return fmt.Errorf("failed to create code file generator: %w", err)
	}

	errorCollector := NewErrorCollector()

	// Generate all files using the code file generator
	errorCollector.Add(codeFileGenerator.GenerateCommonSchemaFile(config))
	errorCollector.Add(codeFileGenerator.GenerateRequestTypesFile(config))
	errorCollector.Add(codeFileGenerator.GenerateResponseSchemaFile(config))
	errorCollector.Add(codeFileGenerator.GenerateTypeScriptTypesFile(config))

	return errorCollector.CombinedError()
}

// createUsageTracker creates maps to track definition usage
func (orchestrator *SchemaOrchestrator) createUsageTracker() (map[string]bool, map[string]bool) {
	requestUsageMap := make(map[string]bool)
	responseUsageMap := make(map[string]bool)

	for _, schemaInfo := range orchestrator.schemas {
		var targetUsageMap map[string]bool
		if schemaInfo.IsRequest {
			targetUsageMap = requestUsageMap
		} else {
			targetUsageMap = responseUsageMap
		}

		// Mark main type as used
		if schemaInfo.MainType != "" {
			targetUsageMap[schemaInfo.MainType] = true
		}

		// Track dependencies
		orchestrator.markDependenciesAsUsed(schemaInfo.Schema.Defs, targetUsageMap)
	}

	return requestUsageMap, responseUsageMap
}

// markDependenciesAsUsed marks all dependencies as used in the usage map
func (orchestrator *SchemaOrchestrator) markDependenciesAsUsed(definitions map[string]SchemaProperty, usageMap map[string]bool) {
	for _, definition := range definitions {
		dependencies := GetDependencies(definition)
		for _, dependency := range dependencies {
			usageMap[dependency] = true
		}
	}
}

// categorizeDefinitionsByUsage categorizes definitions based on usage patterns
func (orchestrator *SchemaOrchestrator) categorizeDefinitionsByUsage(requestUsageMap, responseUsageMap map[string]bool) {
	for definitionName, definition := range orchestrator.allDefinitions {
		usedInRequests := requestUsageMap[definitionName]
		usedInResponses := responseUsageMap[definitionName]

		switch {
		case usedInRequests && usedInResponses:
			orchestrator.commonDefinitions[definitionName] = definition
		case usedInRequests:
			orchestrator.requestDefinitions[definitionName] = definition
		case usedInResponses:
			orchestrator.responseDefinitions[definitionName] = definition
		default:
			// Default to common if usage is unclear
			orchestrator.commonDefinitions[definitionName] = definition
		}
	}
}

// buildRPCMethodMappings builds mappings from type names to RPC methods
func (orchestrator *SchemaOrchestrator) buildRPCMethodMappings() {
	for _, schemaInfo := range orchestrator.schemas {
		if schemaInfo.MainType == "" || schemaInfo.RPCMethod == "" {
			continue
		}

		if schemaInfo.IsRequest {
			orchestrator.requestTypeMappings[schemaInfo.MainType] = schemaInfo.RPCMethod
		} else {
			orchestrator.responseTypeMappings[schemaInfo.MainType] = schemaInfo.RPCMethod
		}
	}
}

// createGeneratorDependencies creates dependencies for the code file generator
func (orchestrator *SchemaOrchestrator) createGeneratorDependencies() *GeneratorDependencies {
	return &GeneratorDependencies{
		RequestDefinitions:   orchestrator.requestDefinitions,
		ResponseDefinitions:  orchestrator.responseDefinitions,
		CommonDefinitions:    orchestrator.commonDefinitions,
		RequestTypeMappings:  orchestrator.requestTypeMappings,
		ResponseTypeMappings: orchestrator.responseTypeMappings,
		DefinitionSorter:     orchestrator.getSortedDefinitionNames,
		EnumNameConverter:    orchestrator.convertRPCMethodToEnumName,
	}
}

func (orchestrator *SchemaOrchestrator) getSortedDefinitionNames(definitions map[string]SchemaProperty) []string {
	dependencyGraph := orchestrator.buildDependencyGraph(definitions)
	return orchestrator.topologicalSort(dependencyGraph, definitions)
}

// buildDependencyGraph builds a dependency graph for definitions
func (orchestrator *SchemaOrchestrator) buildDependencyGraph(definitions map[string]SchemaProperty) map[string][]string {
	dependencyGraph := make(map[string][]string)

	for definitionName, definition := range definitions {
		dependencies := GetDependencies(definition)
		var validDependencies []string

		for _, dependency := range dependencies {
			if _, exists := definitions[dependency]; exists {
				validDependencies = append(validDependencies, dependency)
			}
		}

		dependencyGraph[definitionName] = validDependencies
	}

	return dependencyGraph
}

// topologicalSort performs topological sort on the dependency graph
func (orchestrator *SchemaOrchestrator) topologicalSort(dependencyGraph map[string][]string, definitions map[string]SchemaProperty) []string {
	visitedNodes := make(map[string]bool)
	visitingNodes := make(map[string]bool)
	sortedResult := make([]string, 0, len(definitions))

	var visitNode func(string) bool
	visitNode = func(nodeName string) bool {
		if visitingNodes[nodeName] {
			return true // Circular dependency detected, continue
		}
		if visitedNodes[nodeName] {
			return true
		}

		visitingNodes[nodeName] = true

		// Visit dependencies first
		for _, dependency := range dependencyGraph[nodeName] {
			visitNode(dependency)
		}

		visitingNodes[nodeName] = false
		visitedNodes[nodeName] = true
		sortedResult = append(sortedResult, nodeName)
		return true
	}

	// Get all definition names and sort them for consistent ordering
	definitionNames := make([]string, 0, len(definitions))
	for name := range definitions {
		definitionNames = append(definitionNames, name)
	}
	sort.Strings(definitionNames)

	// Visit each node
	for _, name := range definitionNames {
		visitNode(name)
	}

	return sortedResult
}

func (orchestrator *SchemaOrchestrator) convertRPCMethodToEnumName(rpcMethod string) string {
	methodParts := strings.Split(rpcMethod, "_")
	for i, part := range methodParts {
		methodParts[i] = cases.Title(language.English).String(part)
	}
	return strings.Join(methodParts, "")
}
