import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { KnowledgeManagement } from '../KnowledgeManagement'
import { apiClient } from '../../lib/api'

// Mock the API client
vi.mock('../../lib/api', () => ({
  apiClient: {
    getDocuments: vi.fn(),
    getScrapingJobs: vi.fn(),
    uploadDocument: vi.fn(),
    deleteDocument: vi.fn(),
    createScrapingJob: vi.fn(),
    searchKnowledge: vi.fn(),
  },
}))

// Mock data
const mockDocuments = [
  {
    id: '1',
    title: 'Test Document 1',
    file_name: 'test1.pdf',
    file_size: 1024,
    file_type: 'application/pdf',
    status: 'ready',
    uploaded_at: '2023-01-01T00:00:00Z',
    chunk_count: 5,
  },
  {
    id: '2',
    title: 'Test Document 2',
    file_name: 'test2.pdf',
    file_size: 2048,
    file_type: 'application/pdf',
    status: 'processing',
    uploaded_at: '2023-01-02T00:00:00Z',
    chunk_count: 0,
  },
]

const mockScrapingJobs = [
  {
    id: '1',
    url: 'https://example.com',
    status: 'completed',
    max_depth: 3,
    pages_scraped: 10,
    total_pages: 10,
    created_at: '2023-01-01T00:00:00Z',
    completed_at: '2023-01-01T01:00:00Z',
  },
  {
    id: '2',
    url: 'https://test.com',
    status: 'running',
    max_depth: 2,
    pages_scraped: 5,
    total_pages: 8,
    created_at: '2023-01-02T00:00:00Z',
    completed_at: null,
  },
]

const mockSearchResults = {
  documents: mockDocuments,
  chunks: [
    {
      id: '1',
      document_id: '1',
      content: 'This is a test chunk from document 1',
      similarity: 0.85,
    },
  ],
  query: 'test',
  total_results: 1,
}

describe('KnowledgeManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Setup default mock implementations
    ;(apiClient.getDocuments as any).mockResolvedValue(mockDocuments)
    ;(apiClient.getScrapingJobs as any).mockResolvedValue(mockScrapingJobs)
  })

  describe('Component Rendering', () => {
    it('should render loading state initially', () => {
      render(<KnowledgeManagement projectId="test-project" />)
      expect(screen.getByRole('generic')).toBeInTheDocument()
    })

    it('should render "no project selected" message when projectId is null', () => {
      render(<KnowledgeManagement projectId={null} />)
      expect(screen.getByText('Please select a project to manage knowledge base.')).toBeInTheDocument()
    })

    it('should render knowledge management interface when data is loaded', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Knowledge Base Management')).toBeInTheDocument()
      })
      
      expect(screen.getByText('Documents')).toBeInTheDocument()
      expect(screen.getByText('Web Scraping Jobs')).toBeInTheDocument()
      expect(screen.getByText('Search Knowledge Base')).toBeInTheDocument()
    })
  })

  describe('Data Loading', () => {
    it('should load documents and scraping jobs on mount', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(apiClient.getDocuments).toHaveBeenCalledWith('test-project')
        expect(apiClient.getScrapingJobs).toHaveBeenCalledWith('test-project')
      })
    })

    it('should handle loading errors gracefully', async () => {
      const error = new Error('Failed to load data')
      ;(apiClient.getDocuments as any).mockRejectedValue(error)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText(/Failed to load knowledge base data/)).toBeInTheDocument()
      })
    })

    it('should reload data when projectId changes', async () => {
      const { rerender } = render(<KnowledgeManagement projectId="project-1" />)
      
      await waitFor(() => {
        expect(apiClient.getDocuments).toHaveBeenCalledWith('project-1')
      })
      
      rerender(<KnowledgeManagement projectId="project-2" />)
      
      await waitFor(() => {
        expect(apiClient.getDocuments).toHaveBeenCalledWith('project-2')
      })
    })
  })

  describe('Document Management', () => {
    it('should display uploaded documents', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Document 1')).toBeInTheDocument()
        expect(screen.getByText('Test Document 2')).toBeInTheDocument()
        expect(screen.getByText('test1.pdf')).toBeInTheDocument()
        expect(screen.getByText('test2.pdf')).toBeInTheDocument()
      })
    })

    it('should show document status icons', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        const readyIcons = screen.getAllByTestId('check-circle-icon')
        const processingIcons = screen.getAllByTestId('loader-icon')
        expect(readyIcons.length).toBeGreaterThan(0)
        expect(processingIcons.length).toBeGreaterThan(0)
      })
    })

    it('should handle document upload via drag and drop', async () => {
      ;(apiClient.uploadDocument as any).mockResolvedValue(mockDocuments[0])
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const dropZone = screen.getByText(/Drag and drop PDF files here/)
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })
      
      await waitFor(() => {
        expect(apiClient.uploadDocument).toHaveBeenCalledWith('test-project', file)
      })
    })

    it('should handle document upload via file input', async () => {
      ;(apiClient.uploadDocument as any).mockResolvedValue(mockDocuments[0])
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const fileInput = screen.getByDisplayValue('')
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      
      fireEvent.change(fileInput, { target: { files: [file] } })
      
      await waitFor(() => {
        expect(apiClient.uploadDocument).toHaveBeenCalledWith('test-project', file)
      })
    })

    it('should validate file type during upload', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const dropZone = screen.getByText(/Drag and drop PDF files here/)
      const file = new File(['test content'], 'test.txt', { type: 'text/plain' })
      
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })
      
      await waitFor(() => {
        expect(screen.getByText('Please select PDF files only')).toBeInTheDocument()
      })
      
      expect(apiClient.uploadDocument).not.toHaveBeenCalled()
    })

    it('should handle document deletion', async () => {
      ;(apiClient.deleteDocument as any).mockResolvedValue(undefined)
      global.confirm = vi.fn(() => true)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Document 1')).toBeInTheDocument()
      })
      
      const deleteButtons = screen.getAllByText('Delete')
      fireEvent.click(deleteButtons[0])
      
      await waitFor(() => {
        expect(apiClient.deleteDocument).toHaveBeenCalledWith('test-project', '1')
      })
    })

    it('should not delete document if user cancels confirmation', async () => {
      global.confirm = vi.fn(() => false)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Document 1')).toBeInTheDocument()
      })
      
      const deleteButtons = screen.getAllByText('Delete')
      fireEvent.click(deleteButtons[0])
      
      expect(apiClient.deleteDocument).not.toHaveBeenCalled()
    })
  })

  describe('Web Scraping', () => {
    it('should display scraping jobs', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('https://example.com')).toBeInTheDocument()
        expect(screen.getByText('https://test.com')).toBeInTheDocument()
        expect(screen.getByText('10/10 pages')).toBeInTheDocument()
        expect(screen.getByText('5/8 pages')).toBeInTheDocument()
      })
    })

    it('should show scraping form when "Add Web Source" is clicked', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      expect(screen.getByLabelText('Website URL')).toBeInTheDocument()
      expect(screen.getByLabelText('Max Crawl Depth')).toBeInTheDocument()
    })

    it('should create scraping job with valid URL', async () => {
      ;(apiClient.createScrapingJob as any).mockResolvedValue(mockScrapingJobs[0])
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      const urlInput = screen.getByLabelText('Website URL')
      const depthInput = screen.getByLabelText('Max Crawl Depth')
      const startButton = screen.getByText('Start Scraping')
      
      fireEvent.change(urlInput, { target: { value: 'https://example.com' } })
      fireEvent.change(depthInput, { target: { value: '2' } })
      fireEvent.click(startButton)
      
      await waitFor(() => {
        expect(apiClient.createScrapingJob).toHaveBeenCalledWith('test-project', {
          url: 'https://example.com',
          max_depth: 2,
        })
      })
    })

    it('should validate URL before creating scraping job', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      const startButton = screen.getByText('Start Scraping')
      fireEvent.click(startButton)
      
      await waitFor(() => {
        expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument()
      })
      
      expect(apiClient.createScrapingJob).not.toHaveBeenCalled()
    })
  })

  describe('Knowledge Search', () => {
    it('should perform search when search button is clicked', async () => {
      ;(apiClient.searchKnowledge as any).mockResolvedValue(mockSearchResults)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search knowledge base...')).toBeInTheDocument()
      })
      
      const searchInput = screen.getByPlaceholderText('Search knowledge base...')
      const searchButton = screen.getByText('Search')
      
      fireEvent.change(searchInput, { target: { value: 'test query' } })
      fireEvent.click(searchButton)
      
      await waitFor(() => {
        expect(apiClient.searchKnowledge).toHaveBeenCalledWith('test-project', 'test query')
      })
    })

    it('should display search results', async () => {
      ;(apiClient.searchKnowledge as any).mockResolvedValue(mockSearchResults)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search knowledge base...')).toBeInTheDocument()
      })
      
      const searchInput = screen.getByPlaceholderText('Search knowledge base...')
      const searchButton = screen.getByText('Search')
      
      fireEvent.change(searchInput, { target: { value: 'test query' } })
      fireEvent.click(searchButton)
      
      await waitFor(() => {
        expect(screen.getByText('Search Results')).toBeInTheDocument()
        expect(screen.getByText('This is a test chunk from document 1')).toBeInTheDocument()
        expect(screen.getByText('Similarity: 85%')).toBeInTheDocument()
      })
    })

    it('should not search with empty query', async () => {
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search knowledge base...')).toBeInTheDocument()
      })
      
      const searchButton = screen.getByText('Search')
      fireEvent.click(searchButton)
      
      expect(apiClient.searchKnowledge).not.toHaveBeenCalled()
    })
  })

  describe('Error Handling', () => {
    it('should handle upload errors', async () => {
      const error = new Error('Upload failed')
      ;(apiClient.uploadDocument as any).mockRejectedValue(error)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const dropZone = screen.getByText(/Drag and drop PDF files here/)
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })
      
      await waitFor(() => {
        expect(screen.getByText('Failed to upload documents')).toBeInTheDocument()
      })
    })

    it('should handle deletion errors', async () => {
      const error = new Error('Delete failed')
      ;(apiClient.deleteDocument as any).mockRejectedValue(error)
      global.confirm = vi.fn(() => true)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Document 1')).toBeInTheDocument()
      })
      
      const deleteButtons = screen.getAllByText('Delete')
      fireEvent.click(deleteButtons[0])
      
      await waitFor(() => {
        expect(screen.getByText('Failed to delete document')).toBeInTheDocument()
      })
    })

    it('should handle scraping job creation errors', async () => {
      const error = new Error('Scraping failed')
      ;(apiClient.createScrapingJob as any).mockRejectedValue(error)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      const urlInput = screen.getByLabelText('Website URL')
      const startButton = screen.getByText('Start Scraping')
      
      fireEvent.change(urlInput, { target: { value: 'https://example.com' } })
      fireEvent.click(startButton)
      
      await waitFor(() => {
        expect(screen.getByText('Failed to create scraping job')).toBeInTheDocument()
      })
    })

    it('should handle search errors', async () => {
      const error = new Error('Search failed')
      ;(apiClient.searchKnowledge as any).mockRejectedValue(error)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search knowledge base...')).toBeInTheDocument()
      })
      
      const searchInput = screen.getByPlaceholderText('Search knowledge base...')
      const searchButton = screen.getByText('Search')
      
      fireEvent.change(searchInput, { target: { value: 'test query' } })
      fireEvent.click(searchButton)
      
      await waitFor(() => {
        expect(screen.getByText('Search failed')).toBeInTheDocument()
      })
    })
  })

  describe('Success Messages', () => {
    it('should show success message after successful upload', async () => {
      ;(apiClient.uploadDocument as any).mockResolvedValue(mockDocuments[0])
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const dropZone = screen.getByText(/Drag and drop PDF files here/)
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })
      
      await waitFor(() => {
        expect(screen.getByText('Successfully uploaded 1 document(s)')).toBeInTheDocument()
      })
    })

    it('should show success message after successful deletion', async () => {
      ;(apiClient.deleteDocument as any).mockResolvedValue(undefined)
      global.confirm = vi.fn(() => true)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Document 1')).toBeInTheDocument()
      })
      
      const deleteButtons = screen.getAllByText('Delete')
      fireEvent.click(deleteButtons[0])
      
      await waitFor(() => {
        expect(screen.getByText('Document deleted successfully')).toBeInTheDocument()
      })
    })

    it('should show success message after successful scraping job creation', async () => {
      ;(apiClient.createScrapingJob as any).mockResolvedValue(mockScrapingJobs[0])
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      const urlInput = screen.getByLabelText('Website URL')
      const startButton = screen.getByText('Start Scraping')
      
      fireEvent.change(urlInput, { target: { value: 'https://example.com' } })
      fireEvent.click(startButton)
      
      await waitFor(() => {
        expect(screen.getByText('Scraping job created successfully')).toBeInTheDocument()
      })
    })
  })

  describe('Loading States', () => {
    it('should show upload loading state', async () => {
      let resolveUpload: any
      const uploadPromise = new Promise((resolve) => {
        resolveUpload = resolve
      })
      ;(apiClient.uploadDocument as any).mockReturnValue(uploadPromise)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Documents')).toBeInTheDocument()
      })
      
      const dropZone = screen.getByText(/Drag and drop PDF files here/)
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })
      
      expect(screen.getByText('Uploading...')).toBeInTheDocument()
      
      resolveUpload(mockDocuments[0])
    })

    it('should show scraping loading state', async () => {
      let resolveJob: any
      const jobPromise = new Promise((resolve) => {
        resolveJob = resolve
      })
      ;(apiClient.createScrapingJob as any).mockReturnValue(jobPromise)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByText('Add Web Source')).toBeInTheDocument()
      })
      
      fireEvent.click(screen.getByText('Add Web Source'))
      
      const urlInput = screen.getByLabelText('Website URL')
      const startButton = screen.getByText('Start Scraping')
      
      fireEvent.change(urlInput, { target: { value: 'https://example.com' } })
      fireEvent.click(startButton)
      
      expect(screen.getByText('Creating scraping job...')).toBeInTheDocument()
      
      resolveJob(mockScrapingJobs[0])
    })

    it('should show search loading state', async () => {
      let resolveSearch: any
      const searchPromise = new Promise((resolve) => {
        resolveSearch = resolve
      })
      ;(apiClient.searchKnowledge as any).mockReturnValue(searchPromise)
      
      render(<KnowledgeManagement projectId="test-project" />)
      
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search knowledge base...')).toBeInTheDocument()
      })
      
      const searchInput = screen.getByPlaceholderText('Search knowledge base...')
      const searchButton = screen.getByText('Search')
      
      fireEvent.change(searchInput, { target: { value: 'test query' } })
      fireEvent.click(searchButton)
      
      expect(screen.getByText('Searching...')).toBeInTheDocument()
      
      resolveSearch(mockSearchResults)
    })
  })
})
