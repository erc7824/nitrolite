package main

import (
	"fmt"
	"sort"
	"strings"
)

type ZodGenerator struct {
	schemas                map[string]SchemaInfo
	allDefinitions         map[string]SchemaProperty
	commonDefinitions      map[string]SchemaProperty
	requestDefinitions     map[string]SchemaProperty
	responseDefinitions    map[string]SchemaProperty
	requestTypeMappings    map[string]string // typeName -> rpcMethod
	responseTypeMappings   map[string]string // typeName -> rpcMethod
	unifiedGenerator       *UnifiedGenerator
}

func NewZodGenerator() *ZodGenerator {
	return &ZodGenerator{
		schemas:              make(map[string]SchemaInfo),
		allDefinitions:       make(map[string]SchemaProperty),
		commonDefinitions:    make(map[string]SchemaProperty),
		requestDefinitions:   make(map[string]SchemaProperty),
		responseDefinitions:  make(map[string]SchemaProperty),
		requestTypeMappings:  make(map[string]string),
		responseTypeMappings: make(map[string]string),
	}
}

func (g *ZodGenerator) LoadSchemas(requestDirectoryPath, responseDirectoryPath string) error {
	loadedSchemas, allDefinitions, err := LoadSchemas(requestDirectoryPath, responseDirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	g.schemas = loadedSchemas
	g.allDefinitions = allDefinitions
	return nil
}

func (g *ZodGenerator) CategorizeDefinitions() {
	requestUsageMap, responseUsageMap := g.createUsageTracker()
	g.categorizeDefinitionsByUsage(requestUsageMap, responseUsageMap)
	g.buildRPCMethodMappings()
}

func (g *ZodGenerator) GenerateAllFiles(schemaDirectoryPath string, sdkRootDirectoryPath string) error {
	config, err := NewGenerationConfig(schemaDirectoryPath, sdkRootDirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to create generation config: %w", err)
	}

	dependencies := g.createGeneratorDependencies()
	unifiedGenerator, err := NewUnifiedGenerator(dependencies)
	if err != nil {
		return fmt.Errorf("failed to create unified generator: %w", err)
	}

	errorCollector := NewErrorCollector()
	
	// Generate all files using the unified generator
	errorCollector.Add(unifiedGenerator.GenerateCommonSchemaFile(config))
	errorCollector.Add(unifiedGenerator.GenerateRequestSchemaFile(config))
	errorCollector.Add(unifiedGenerator.GenerateResponseSchemaFile(config))
	errorCollector.Add(unifiedGenerator.GenerateTypeScriptTypesFile(config))

	return errorCollector.CombinedError()
}

// createUsageTracker creates maps to track definition usage
func (g *ZodGenerator) createUsageTracker() (map[string]bool, map[string]bool) {
	requestUsageMap := make(map[string]bool)
	responseUsageMap := make(map[string]bool)

	for _, schemaInfo := range g.schemas {
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
		g.markDependenciesAsUsed(schemaInfo.Schema.Defs, targetUsageMap)
	}

	return requestUsageMap, responseUsageMap
}

// markDependenciesAsUsed marks all dependencies as used in the usage map
func (g *ZodGenerator) markDependenciesAsUsed(definitions map[string]SchemaProperty, usageMap map[string]bool) {
	for _, definition := range definitions {
		dependencies := GetDependencies(definition)
		for _, dependency := range dependencies {
			usageMap[dependency] = true
		}
	}
}

// categorizeDefinitionsByUsage categorizes definitions based on usage patterns
func (g *ZodGenerator) categorizeDefinitionsByUsage(requestUsageMap, responseUsageMap map[string]bool) {
	for definitionName, definition := range g.allDefinitions {
		usedInRequests := requestUsageMap[definitionName]
		usedInResponses := responseUsageMap[definitionName]

		switch {
		case usedInRequests && usedInResponses:
			g.commonDefinitions[definitionName] = definition
		case usedInRequests:
			g.requestDefinitions[definitionName] = definition
		case usedInResponses:
			g.responseDefinitions[definitionName] = definition
		default:
			// Default to common if usage is unclear
			g.commonDefinitions[definitionName] = definition
		}
	}
}

// buildRPCMethodMappings builds mappings from type names to RPC methods
func (g *ZodGenerator) buildRPCMethodMappings() {
	for _, schemaInfo := range g.schemas {
		if schemaInfo.MainType == "" || schemaInfo.RPCMethod == "" {
			continue
		}

		if schemaInfo.IsRequest {
			g.requestTypeMappings[schemaInfo.MainType] = schemaInfo.RPCMethod
		} else {
			g.responseTypeMappings[schemaInfo.MainType] = schemaInfo.RPCMethod
		}
	}
}

// createGeneratorDependencies creates dependencies for the unified generator
func (g *ZodGenerator) createGeneratorDependencies() *GeneratorDependencies {
	return &GeneratorDependencies{
		RequestDefinitions:   g.requestDefinitions,
		ResponseDefinitions:  g.responseDefinitions,
		CommonDefinitions:    g.commonDefinitions,
		RequestTypeMappings:  g.requestTypeMappings,
		ResponseTypeMappings: g.responseTypeMappings,
		DefinitionSorter:     g.getSortedDefinitionNames,
		EnumNameConverter:    g.convertRPCMethodToEnumName,
	}
}

func (g *ZodGenerator) getSortedDefinitionNames(definitions map[string]SchemaProperty) []string {
	dependencyGraph := g.buildDependencyGraph(definitions)
	return g.topologicalSort(dependencyGraph, definitions)
}

// buildDependencyGraph builds a dependency graph for definitions
func (g *ZodGenerator) buildDependencyGraph(definitions map[string]SchemaProperty) map[string][]string {
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
func (g *ZodGenerator) topologicalSort(dependencyGraph map[string][]string, definitions map[string]SchemaProperty) []string {
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

func (g *ZodGenerator) convertRPCMethodToEnumName(rpcMethod string) string {
	methodParts := strings.Split(rpcMethod, "_")
	for i, part := range methodParts {
		methodParts[i] = strings.Title(part)
	}
	return strings.Join(methodParts, "")
}

