import React from 'react';

interface Param {
  name: string;
  type: string;
}

interface MethodCardProps {
  name: string;
  description: string;
  params?: Param[];
  returns?: string;
  example?: string;
}

const MethodCard: React.FC<MethodCardProps> = ({
  name,
  description,
  params = [],
  returns,
  example,
}) => {
  return (
    <article className="rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow p-4 bg-white dark:bg-gray-800 dark:border-gray-700 h-full flex flex-col">
      <header className="mb-4">
        <h3 className="text-xl font-mono font-bold text-gray-900 dark:text-gray-100">
          {name}
        </h3>
        <p className="text-gray-600 dark:text-gray-300 mt-2">
          {description}
        </p>
      </header>

      <div className="flex-grow">
        {params.length > 0 && (
          <section className="mb-4">
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              Parameters
            </h4>
            <ul className="space-y-1">
              {params.map((param, index) => (
                <li key={index} className="text-sm">
                  <span className="font-medium text-gray-800 dark:text-gray-200">
                    {param.name}:
                  </span>{' '}
                  <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">
                    {param.type}
                  </code>
                </li>
              ))}
            </ul>
          </section>
        )}

        {returns && (
          <section className="mb-4">
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              Returns
            </h4>
            <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">
              {returns}
            </code>
          </section>
        )}
      </div>

      {example && (
        <section className="mt-2">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
            Example
          </h4>
          <pre className="bg-gray-100 dark:bg-gray-700 p-2 rounded-md overflow-x-auto text-sm">
            <code>{example}</code>
          </pre>
        </section>
      )}
    </article>
  );
};

export default MethodCard;