import { Component, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: any) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-lg p-8 max-w-2xl w-full">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                <span className="text-2xl">⚠️</span>
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Something went wrong</h1>
                <p className="text-gray-600">The application encountered an error</p>
              </div>
            </div>

            <div className="bg-red-50 border border-red-200 rounded p-4 mb-4">
              <p className="font-semibold text-red-800 mb-2">Error Details:</p>
              <code className="text-sm text-red-700 break-all">
                {this.state.error?.message || 'Unknown error'}
              </code>
            </div>

            <details className="mb-4">
              <summary className="cursor-pointer text-sm text-gray-600 hover:text-gray-800 font-medium">
                View Stack Trace
              </summary>
              <pre className="mt-2 p-3 bg-gray-100 rounded text-xs overflow-auto max-h-60">
                {this.state.error?.stack || 'No stack trace available'}
              </pre>
            </details>

            <div className="flex gap-3">
              <button
                onClick={() => window.location.reload()}
                className="flex-1 bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded font-medium transition"
              >
                Reload Page
              </button>
              <button
                onClick={() => this.setState({ hasError: false, error: null })}
                className="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-800 px-6 py-3 rounded font-medium transition"
              >
                Try Again
              </button>
            </div>

            <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded text-sm">
              <p className="text-blue-800">
                <strong>Tip:</strong> Check the browser console (F12) for more details, or try disconnecting and reconnecting your wallet.
              </p>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
