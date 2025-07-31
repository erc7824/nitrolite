import React from "react";
import Link from "@docusaurus/Link";
import CodeBlock from "@theme/CodeBlock";

// Common custom types that should link to the types documentation
// Map type names to their corresponding anchors in the types.md page - must match the actual heading IDs
const TYPE_ANCHORS = {
    CreateChannelParams: "#2-channel-creation",
    CloseChannelParams: "#4-channel-closing",
    CheckpointChannelParams: "#3-channel-operations",
    ChallengeChannelParams: "#3-channel-operations",
    ResizeChannelParams: "#3-channel-operations",
    ChannelId: "#channelid",
    State: "#state",
    StateIntent: "#stateintent",
    Allocation: "#allocation",
    Signature: "#signature",
    AccountInfo: "#accountinfo",
    Hash: "#core-types",
    PreparedTransaction: "#preparedtransaction",
    ContractAddresses: "#contractaddresses",
    NitroliteClientConfig: "#nitroliteclientconfig",
};

// Get all type names for checking
const CUSTOM_TYPES = Object.keys(TYPE_ANCHORS);

interface Param {
    name: string;
    type: string;
}

interface MethodDetailsProps {
    name: string;
    description: string;
    params?: Param[];
    returns?: string;
    example?: string;
}

const MethodDetails: React.FC<MethodDetailsProps> = ({ name, description, params = [], returns, example }) => {
    // Function to render a type with a link if it's a custom type
    const renderType = (type: string) => {
        // Extract the base type from arrays, promises, etc.
        let baseType = type;

        // Extract from Promise<Type>
        const promiseMatch = type.match(/Promise<(.+)>/);
        if (promiseMatch) baseType = promiseMatch[1];

        // Handle complex nested types
        baseType = baseType.replace(/\{.*\}/, ""); // Remove object literals
        baseType = baseType.replace(/\[.*\]/, ""); // Remove array literals

        // Check if any custom type is found in the type string
        const foundCustomType = CUSTOM_TYPES.find((customType) => baseType.includes(customType));

        if (foundCustomType) {
            const anchor = TYPE_ANCHORS[foundCustomType];
            return (
                <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">
                    {type.split(foundCustomType).map((part, i, parts) => (
                        <React.Fragment key={i}>
                            {part}
                            {i < parts.length - 1 && (
                                <Link
                                    to={`./types${anchor}`}
                                    className="text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
                                >
                                    {foundCustomType}
                                </Link>
                            )}
                        </React.Fragment>
                    ))}
                </code>
            );
        }

        return <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">{type}</code>;
    };

    return (
        <details className="mb-6 rounded-lg border border-gray-200 shadow-sm bg-white dark:bg-gray-800 dark:border-gray-700 p-4">
            <summary className="font-mono font-bold text-lg cursor-pointer marker:text-transparent flex items-center">{name}</summary>
            <div className="pt-4 pl-6">
                <p className="text-gray-600 dark:text-gray-300 mb-4">{description}</p>

                {params.length > 0 && (
                    <div className="mb-4">
                        <p className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Parameters</p>
                        <ul className="pl-4 space-y-1">
                            {params.map((param, index) => (
                                <li key={index} className="text-sm flex items-baseline">
                                    <span className="font-medium text-gray-800 dark:text-gray-200 mr-2">{param.name}:</span>
                                    {renderType(param.type)}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                {returns && (
                    <p className="mb-4 flex items-baseline">
                        <strong className="text-sm font-semibold text-gray-700 dark:text-gray-300 mr-2">Returns:</strong>
                        {renderType(returns)}
                    </p>
                )}

                {example && (
                    <div className="mb-2">
                        <p className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Example</p>
                        <div className="method-example-code">
                            <CodeBlock language="typescript" className="method-code-block">
                                {example}
                            </CodeBlock>
                        </div>
                    </div>
                )}
            </div>
        </details>
    );
};

// Add styles for the summary icon rotation and code formatting
const styles = `
  details[open] .summary-icon {
    transform: rotate(90deg);
  }
  
  .method-example-code pre {
    tab-size: 2;
    -moz-tab-size: 2;
    white-space: pre !important;
  }
  
  .method-example-code code {
    white-space: pre !important;
  }
`;

// Add the styles to the document
if (typeof document !== "undefined") {
    const styleElement = document.createElement("style");
    styleElement.textContent = styles;
    document.head.appendChild(styleElement);
}

export default MethodDetails;
