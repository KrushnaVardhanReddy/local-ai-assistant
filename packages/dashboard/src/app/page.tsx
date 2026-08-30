import React from 'react';

export default function DashboardHome() {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col font-sans">
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center">
            <h1 className="text-xl font-semibold text-gray-900">
              Interview Integrity Dashboard
            </h1>
          </div>
          <div className="flex items-center space-x-4">
            <span className="text-sm text-gray-500">Reviewer Mode</span>
            <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-700 font-bold">
              R
            </div>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 w-full">
        <div className="bg-white rounded-lg shadow border border-gray-200 p-6 sm:p-10 h-full flex flex-col">
          <div className="mb-6 border-b border-gray-200 pb-4">
            <h2 className="text-2xl font-bold text-gray-900">Active Sessions</h2>
            <p className="text-gray-500 mt-1">
              Monitor real-time integrity status and post-interview timelines.
            </p>
          </div>

          <div className="flex-1 flex flex-col items-center justify-center text-center border-2 border-dashed border-gray-300 rounded-lg p-12 bg-gray-50">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No active sessions</h3>
            <p className="mt-1 text-sm text-gray-500">
              Waiting for interview sessions to begin...
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
