import Editor from '@monaco-editor/react'

interface SQLViewerProps {
  code: string
  height?: string
  readOnly?: boolean
}

export default function SQLViewer({ code, height = '400px', readOnly = true }: SQLViewerProps) {
  return (
    <div className="border border-gray-300 rounded-md overflow-hidden">
      <Editor
        height={height}
        defaultLanguage="sql"
        value={code}
        theme="vs-dark"
        options={{
          readOnly,
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          fontSize: 14,
          lineNumbers: 'on',
          automaticLayout: true,
          wordWrap: 'on',
          formatOnPaste: true,
          formatOnType: true,
        }}
      />
    </div>
  )
}
