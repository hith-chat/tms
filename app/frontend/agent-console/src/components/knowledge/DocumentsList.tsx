import { FileText, Trash2, CheckCircle, Loader, XCircle, AlertCircle } from 'lucide-react'
import { KnowledgeDocument } from '../../lib/api'

interface DocumentsListProps {
  documents: KnowledgeDocument[]
  onDeleteDocument: (documentId: string) => void
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'ready':
    case 'completed':
      return <CheckCircle className="h-4 w-4 text-green-500" />
    case 'processing':
    case 'running':
      return <Loader className="h-4 w-4 text-blue-500 animate-spin" />
    case 'failed':
    case 'error':
      return <XCircle className="h-4 w-4 text-red-500" />
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />
  }
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function DocumentsList({ documents, onDeleteDocument }: DocumentsListProps) {
  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="space-y-4">
        <div>
          <h3 className="font-medium text-foreground">Uploaded Documents</h3>
          <p className="text-sm text-muted-foreground mt-1">Manage your uploaded PDF documents</p>
        </div>

        <div className="border rounded-lg overflow-hidden">
          <div className="bg-muted/50 px-4 py-3 border-b">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-foreground">Documents</span>
              <span className="text-sm text-muted-foreground">{documents.length} files</span>
            </div>
          </div>
          <div className="divide-y">
            {documents.map((doc) => (
              <div key={doc.id} className="flex items-center justify-between p-4">
                <div className="flex items-center space-x-3">
                  <FileText className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium text-foreground">{doc.filename}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatFileSize(doc.file_size)} • {new Date(doc.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(doc.status)}
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        doc.status === 'ready'
                          ? 'bg-green-100 text-green-800'
                          : doc.status === 'processing'
                          ? 'bg-blue-100 text-blue-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {doc.status}
                    </span>
                  </div>
                  <button
                    onClick={() => onDeleteDocument(doc.id)}
                    className="inline-flex items-center px-2 py-1 text-xs font-medium text-muted-foreground hover:text-destructive transition-colors"
                    aria-label="Delete document"
                  >
                    <Trash2 className="h-3 w-3 mr-1" />
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
