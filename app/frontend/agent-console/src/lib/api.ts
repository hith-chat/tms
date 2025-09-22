import axios, { AxiosInstance, AxiosResponse } from 'axios'
import type {
  ChatWidget,
  ChatSession,
  ChatMessage,
  CreateChatWidgetRequest,
  UpdateChatWidgetRequest,
  SendChatMessageRequest,
  AssignChatSessionRequest,
  ChatSessionFilters,
  AIMetrics
} from '../types/chat'
import type { Notification, NotificationCount, HowlingAlarm, AlarmStats } from '../types/notifications'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/v1'

// Handoff-related interfaces
export interface HandoffResponse {
  success: boolean
  session_id: string
  agent_id: string
  tenant_id: string
  accepted_at?: string
  declined_at?: string
  message: string
}

export interface HandoffStatus {
  session_id: string
  status: 'pending' | 'accepted' | 'declined' | 'expired'
  assigned_agent_id?: string
  handoff_reason: string
  requested_at: string
  expires_at?: string
}

export interface User {
  id: string
  email: string
  name: string
  role: string
  tenant_id: string
  current_project_id?: string
}

export interface Project {
  id: string
  tenant_id: string
  key: string
  name: string
  status?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: User
  projects?: Project[]
}

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
}

export interface TicketsResponse {
  tickets: Ticket[]
  next_cursor?: string
}

export interface Ticket {
  id: string
  number: number
  subject: string
  status: 'new' | 'open' | 'pending' | 'resolved' | 'closed'
  priority: 'low' | 'normal' | 'high' | 'urgent'
  type: 'question' | 'incident' | 'problem' | 'task'
  source: 'web' | 'email' | 'api' | 'phone' | 'chat'
  customer_id: string
  customer? : {
    id: string
    name: string
    email: string
  }
  assignee_agent_id?: string
  tenant_id: string
  project_id: string
  created_at: string
  updated_at: string
  assigned_agent?: {
    id: string
    name: string
    email: string
  }
}

export interface CreateTicketRequest {
  subject: string
  initial_message?: string
  requester_name: string
  requester_email: string
  status?: 'new' | 'open' | 'pending' | 'resolved' | 'closed'
  priority: 'low' | 'normal' | 'high' | 'urgent'
  type: 'question' | 'incident' | 'problem' | 'task'
  source: 'web' | 'email' | 'api' | 'phone' | 'chat'
  customer_id?: string
}

export interface EmailSettings {
  // SMTP Configuration
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_encryption: 'tls' | 'ssl' | 'none'
  
  // IMAP Configuration
  imap_host: string
  imap_port: number
  imap_username: string
  imap_password: string
  imap_encryption: 'tls' | 'ssl' | 'none'
  imap_folder: string
  
  // Email Settings
  from_email: string
  from_name: string
  enable_email_notifications: boolean
  enable_email_to_ticket: boolean
}

export interface DnsMetaData {
  dns_record: string
  dns_value: string
}

export interface DomainValidation {
  id: string
  domain: string
  status: 'pending' | 'verified' | 'failed'
  validation_token?: string
  metadata: DnsMetaData
  verification_proof?: string
  file_name?: string
  file_content?: string
  verified_at?: string
  created_at: string
  updated_at: string
  project_id?: string
  project_name?: string
}

export interface BrandingSettings {
  company_name: string
  about: string
  logo_url: string
  support_url: string
  primary_color: string
  accent_color: string
  secondary_color: string
  custom_css: string
  favicon_url: string
  header_logo_height: number
  enable_custom_branding: boolean
}

export interface AutomationSettings {
  enable_auto_assignment: boolean
  assignment_strategy: string
  max_tickets_per_agent: number
  enable_escalation: boolean
  escalation_threshold_hours: number
  enable_auto_reply: boolean
  auto_reply_template: string
}

export interface AboutMeSettings {
  content: string
}

// Knowledge Management Types
export interface KnowledgeDocument {
  id: string
  tenant_id: string
  project_id: string
  filename: string
  content_type: string
  file_size: number
  file_path: string
  status: 'processing' | 'ready' | 'error'
  error_message?: string
  created_at: string
  updated_at: string
}

export interface KnowledgeScrapingJob {
  id: string
  tenant_id: string
  project_id: string
  url: string
  max_depth: number
  status: 'pending' | 'running' | 'awaiting_selection' | 'indexing' | 'completed' | 'failed' | 'cancelled'
  pages_scraped: number
  total_pages: number
  error_message?: string
  started_at?: string
  completed_at?: string
  staging_file_path?: string
  selected_links: string[]
  indexing_started_at?: string
  indexing_completed_at?: string
  created_at: string
  updated_at: string
}

export interface ScrapedLinkPreview {
  url: string
  title?: string
  depth: number
  token_count: number
  content_preview?: string
  selected: boolean
}

export interface ScrapingJobLinksResponse {
  links: ScrapedLinkPreview[]
  maxSelectableLinks: number
}

export interface SelectScrapingLinksResponse {
  selected_count: number
  message: string
  maxSelectableLinks: number
}

export interface CreateScrapingJobRequest {
  url: string
  max_depth: number
}

export interface KnowledgeSearchResult {
  documents: KnowledgeDocument[]
  scraped_pages: {
    id: string
    url: string
    title?: string
    content: string
    scraped_at: string
  }[]
  total_results: number
}

export interface UpdateTicketRequest {
  subject?: string
  description?: string
  status?: 'new' | 'open' | 'pending' | 'resolved' | 'closed'
  priority?: 'low' | 'normal' | 'high' | 'urgent'
  assignee_agent_id?: string
}

export interface Message {
  id: string
  ticket_id: string
  author_type: 'customer' | 'agent' | 'system'
  author_id: string
  body: string
  is_private: boolean
  created_at: string,
  user_info: {
    id: string
    name: string
    email: string
  }
  attachments?: {
    id: string
    filename: string
    content_type: string
    size: number
    url: string
  }[]
}

export interface MessagesResponse {
  messages: Message[]
  next_cursor?: string
}

export interface CreateMessageRequest {
  content: string
  attachments?: File[]
}

export interface ReassignTicketRequest {
  assignee_agent_id?: string
  note?: string
}

export interface CustomerValidationResult {
  success: boolean
  message: string
  smtp_configured: boolean
  otp_sent?: boolean
}

export interface MagicLinkResult {
  success: boolean
  message: string
  smtp_configured: boolean
  link_sent?: boolean
}

export interface EmailInbox {
  id: string
  message_id: string
  thread_id?: string
  from_address: string
  from_name?: string
  to_addresses: string[]
  subject: string
  snippet?: string
  body_text?: string
  body_html?: string
  sent_at: string
  received_at: string
  is_read: boolean
  is_reply: boolean
  has_attachments: boolean
  attachment_count: number
  is_converted_to_ticket: boolean
  ticket_id?: string
  mailbox_address: string
}

export interface EmailInboxResponse {
  emails: EmailInbox[]
  total: number
}

export interface EmailFilter {
  search?: string
  mailbox?: string
  is_read?: boolean
  is_reply?: boolean
  has_attachments?: boolean
  from_date?: string
  to_date?: string
  thread_id?: string
  page?: number
  limit?: number
}

export interface ConvertToTicketRequest {
  type: string
  priority: string
}

export interface EmailConnector {
  id: string
  project_id: string
  name: string
  type: 'inbound_imap' | 'outbound_smtp'
  imap_host?: string
  imap_port?: number
  imap_username?: string
  smtp_host: string
  smtp_port: number
  smtp_username: string
  is_validated: boolean
  validation_status: 'pending' | 'validating' | 'validated' | 'failed'
  created_at: string
  updated_at: string
}

export interface EmailMailbox {
  id: string
  tenant_id: string
  project_id?: string
  address: string
  display_name?: string
  inbound_connector_id: string
  routing_rules: any
  allow_new_ticket: boolean
  created_at: string
  updated_at: string
  // Additional fields for display
  connector_name?: string
  project_name?: string
  default_project_name?: string
}

export interface EmailConnectorRequest {
  name: string
  type: 'inbound_imap' | 'outbound_smtp'
  imap_host?: string
  imap_port?: number
  imap_username?: string
  imap_password?: string
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
}

export interface EmailMailboxRequest {
  address: string
  display_name?: string
  inbound_connector_id: string
  routing_rules?: Array<{
    match: string
    project_id: string
  }>
  allow_new_ticket: boolean
}

export interface ValidateConnectorRequest {
  email: string
}

export interface VerifyOTPRequest {
  email: string
  otp: string
}

export interface ApiKey {
  id: string
  name: string
  key_preview: string
  created_at: string
  last_used?: string
  is_active: boolean
}

export interface Agent {
  id: string
  name: string
  email: string
  created_at: string
  is_active: boolean
  roles?: Array<{
    role: string
    project_id?: string
    project_name?: string
  }>
}

export interface AgentProject {
  id: string
  name: string
  role: string
}

export interface Customer {
  id: string
  tenant_id: string
  email: string
  name: string
  metadata?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface CustomersResponse {
  customers: Customer[]
  next_cursor?: string
}

export interface CustomerFilters {
  email?: string
  search?: string
  cursor?: string
  limit?: number
}

export interface Integration {
  id: string
  name: string
  type: 'slack' | 'jira' | 'calendly' | 'zapier' | 'webhook' | 'custom' | 'microsoft_teams' | 'github' | 'linear' | 'asana' | 'trello' | 'notion' | 'hubspot' | 'salesforce' | 'zendesk' | 'freshdesk' | 'intercom' | 'discord' | 'google_calendar' | 'zoom' | 'stripe' | 'shopify' | 'email' | 'api' | 'chat'
  status: 'active' | 'inactive' | 'error' | 'configuring'
  config: Record<string, any>
  auth_method?: 'oauth' | 'api_key' | 'none'
  auth_data?: Record<string, any>
  tenant_id: string
  project_id: string
  last_sync_at?: string
  last_error?: string
  created_at: string
  updated_at: string
}

export interface IntegrationCategory {
  id: string
  name: string
  display_name: string
  description?: string
  icon?: string
  sort_order: number
  is_active: boolean
  created_at: string
}

export interface IntegrationTemplate {
  id: string
  category_id: string
  type: string
  name: string
  display_name: string
  description?: string
  logo_url?: string
  website_url?: string
  documentation_url?: string
  auth_method: 'oauth' | 'api_key' | 'none'
  config_schema: Record<string, any>
  default_config: Record<string, any>
  supported_events: string[]
  is_featured: boolean
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface IntegrationCategoryWithTemplates extends IntegrationCategory {
  templates: IntegrationTemplate[]
}

export interface IntegrationWithTemplate extends Integration {
  template?: IntegrationTemplate
  category?: IntegrationCategory
}

class APIClient {
  private client: AxiosInstance
  private tenantId: string | null = null
  private projectId: string | null = null

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor to add auth token and build proper URLs
    this.client.interceptors.request.use((config: any) => {
      const token = localStorage.getItem('auth_token')
      const tenantId = localStorage.getItem('tenant_id') || this.tenantId
      const projectId = localStorage.getItem('project_id') || this.projectId

      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }

      // Build the correct URL structure based on the endpoint
      if (config.url && tenantId) {
        // Auth endpoints: /tenants/{tenant_id}/auth/*
        if (config.url.includes('/auth/') && !config.url.includes('/tenants/')) {
          config.url = `/tenants/${tenantId}${config.url}`
        }
        // Tenant-level endpoints: /tenants/{tenant_id}/* (projects, agents, api-keys at tenant level)
        else if ((
          config.url.startsWith('/projects') || 
          config.url.startsWith('/agents') ||
          config.url.startsWith('/customers')
        ) && !config.url.includes('/tenants/')) {
          config.url = `/tenants/${tenantId}${config.url}`
        }
        // Project-scoped endpoints: /tenants/{tenant_id}/projects/{project_id}/* (tickets, integrations, email)
        else if (projectId && (
          config.url.startsWith('/tickets') || 
          config.url.startsWith('/integrations') ||
          config.url.startsWith('/email') || 
          config.url.startsWith('/settings') ||
          config.url.startsWith('/analytics') ||
          config.url.startsWith('/chat') ||
          config.url.startsWith('/notifications') ||
          config.url.startsWith('/knowledge') ||
          config.url.startsWith('/api-keys') ||
          config.url.startsWith('/alarms')
        ) && !config.url.includes('/tenants/')) {
          config.url = `/tenants/${tenantId}/projects/${projectId}${config.url}`
        }
      }
      return config
    })

    // Response interceptor for error handling and token refresh
    this.client.interceptors.response.use(
      (response: any) => response,
      async (error: any) => {
        const originalRequest = error.config

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true

          try {
            // Try to refresh the token
            await this.refreshToken()
            
            // Retry the original request with new token
            const token = localStorage.getItem('auth_token')
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`
            }
            
            return this.client(originalRequest)
          } catch (refreshError) {
            // Refresh failed, redirect to login
            localStorage.removeItem('auth_token')
            localStorage.removeItem('refresh_token')
            localStorage.removeItem('tenant_id')
            localStorage.removeItem('user_data')
            window.location.href = '/login'
            return Promise.reject(refreshError)
          }
        }

        return Promise.reject(error)
      }
    )
  }

  setTenantId(tenantId: string) {
    this.tenantId = tenantId
    localStorage.setItem('tenant_id', tenantId)
  }

  setProjectId(projectId: string) {
    this.projectId = projectId
    localStorage.setItem('project_id', projectId)
  }

  // Enterprise admin endpoints (cross-tenant)
  async getTenants(): Promise<Array<{id: string, name: string, status: string, region: string, created_at: string, updated_at: string}>> {
    // Use enterprise route that bypasses tenant-scoped interceptor
    const enterpriseClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
      },
    })
    
    const response = await enterpriseClient.get('/enterprise/tenants')
    return response.data.tenants || []
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<LoginResponse> {    
    // Create a separate axios instance for login to avoid the interceptor adding tenant to URL
    const loginClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const response = await loginClient.post<LoginResponse>(`/auth/login`, {
      email: data.email,
      password: data.password
    })
    
    if (response.data.access_token) {
      localStorage.setItem('auth_token', response.data.access_token)
    }
    
    if (response.data.refresh_token) {
      localStorage.setItem('refresh_token', response.data.refresh_token)
    }

    this.setTenantId(response.data.user.tenant_id)

    return response.data
  }

  async refreshToken(): Promise<RefreshTokenResponse> {
    const refreshToken = localStorage.getItem('refresh_token')
    
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const refreshClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const response = await refreshClient.post<RefreshTokenResponse>(`/auth/refresh`, {
      refresh_token: refreshToken
    })

    if (response.data.access_token) {
      localStorage.setItem('auth_token', response.data.access_token)
    }

    if (response.data.refresh_token) {
      localStorage.setItem('refresh_token', response.data.refresh_token)
    }

    return response.data
  }

  async logout(): Promise<void> {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('tenant_id')
    localStorage.removeItem('project_id')
    localStorage.removeItem('user_data')
    this.tenantId = null
    this.projectId = null
  }

  // Signup endpoints
  async signup(data: { email: string; password: string; name: string }): Promise<{ message: string; email: string }> {
    
    // Create a separate axios instance for signup to avoid the interceptor adding tenant to URL
    const signupClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })
    
    const response = await signupClient.post(`/auth/signup`, data)
    return response.data
  }

  async verifySignupOTP(data: { email: string; otp: string }): Promise<LoginResponse> {
    
    // Create a separate axios instance for verification to avoid the interceptor adding tenant to URL
    const verifyClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const response = await verifyClient.post<LoginResponse>(`/auth/verify-signup-otp`, data)
    
    // Store tokens like in login
    localStorage.setItem('auth_token', response.data.access_token)
    localStorage.setItem('refresh_token', response.data.refresh_token)

    this.setTenantId(response.data.user.tenant_id)

    return response.data
  }

  async resendSignupOTP(data: { email: string }): Promise<{ message: string; email: string }> {
    // Create a separate axios instance for resend to avoid the interceptor adding tenant to URL
    const resendClient = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const response = await resendClient.post(`/auth/resend-signup-otp`, data)
    return response.data
  }

  // Project endpoints
  async getProjects(): Promise<Project[]> {
    const response: AxiosResponse<Project[]> = await this.client.get('/projects')
    return response.data
  }

  async getProject(id: string): Promise<Project> {
    const response: AxiosResponse<Project> = await this.client.get(`/projects/${id}`)
    return response.data
  }

  async createProject(data: { key: string; name: string }): Promise<Project> {
    const response: AxiosResponse<Project> = await this.client.post('/projects', data)
    return response.data
  }

  async updateProject(id: string, data: { key: string; name: string; status: string }): Promise<Project> {
    const response: AxiosResponse<Project> = await this.client.put(`/projects/${id}`, data)
    return response.data
  }

  async deleteProject(id: string): Promise<void> {
    await this.client.delete(`/projects/${id}`)
  }

  // Agent endpoints (tenant-scoped)
  async getAgents(): Promise<Agent[]> {
    const response = await this.client.get('/agents')
    return response.data.agents || []
  }

  async createAgent(data: { name: string; email: string; password: string; role: string }): Promise<Agent> {
    const response = await this.client.post('/agents', data)
    return response.data
  }

  async updateAgent(id: string, data: Partial<Agent>): Promise<Agent> {
    const response = await this.client.patch(`/agents/${id}`, data)
    return response.data
  }

  async deleteAgent(id: string): Promise<void> {
    await this.client.delete(`/agents/${id}`)
  }

  // Agent project assignment endpoints
  async getAgentProjects(agentId: string): Promise<AgentProject[]> {
    const response = await this.client.get(`/agents/${agentId}/projects`)
    return response.data.projects || []
  }

  async assignAgentToProject(agentId: string, projectId: string, role: string): Promise<void> {
    await this.client.post(`/agents/${agentId}/projects/${projectId}`, { role })
  }

  async removeAgentFromProject(agentId: string, projectId: string): Promise<void> {
    await this.client.delete(`/agents/${agentId}/projects/${projectId}`)
  }

  // Customer endpoints (tenant-scoped)
  async getCustomers(filters?: CustomerFilters): Promise<CustomersResponse> {
    const params = new URLSearchParams()
    if (filters?.email) params.append('email', filters.email)
    if (filters?.search) params.append('search', filters.search)
    if (filters?.cursor) params.append('cursor', filters.cursor)
    if (filters?.limit) params.append('limit', filters.limit.toString())

    const queryString = params.toString()
    const response: AxiosResponse<CustomersResponse> = await this.client.get(
      `/customers${queryString ? `?${queryString}` : ''}`
    )
    return response.data
  }

  // Ticket endpoints
  async getTickets(): Promise<Ticket[]> {
    const response: AxiosResponse<TicketsResponse> = await this.client.get('/tickets')
    return response.data.tickets
  }

  async getTicket(id: string): Promise<Ticket> {
    const response: AxiosResponse<Ticket> = await this.client.get(`/tickets/${id}`)
    return response.data
  }

  async createTicket(data: CreateTicketRequest): Promise<Ticket> {
    const response: AxiosResponse<Ticket> = await this.client.post('/tickets', data)
    return response.data
  }

  async updateTicket(id: string, data: UpdateTicketRequest): Promise<Ticket> {
    const response: AxiosResponse<Ticket> = await this.client.put(`/tickets/${id}`, data)
    return response.data
  }

  async deleteTicket(id: string): Promise<void> {
    await this.client.delete(`/tickets/${id}`)
  }

  async reassignTicket(id: string, data: ReassignTicketRequest): Promise<Ticket> {
    const response: AxiosResponse<Ticket> = await this.client.post(`/tickets/${id}/reassign`, data)
    return response.data
  }

  async validateCustomer(ticketId: string): Promise<CustomerValidationResult> {
    const response: AxiosResponse<CustomerValidationResult> = await this.client.post(`/tickets/${ticketId}/validate-customer`)
    return response.data
  }

  async sendMagicLinkToCustomer(ticketId: string): Promise<MagicLinkResult> {
    const response: AxiosResponse<MagicLinkResult> = await this.client.post(`/tickets/${ticketId}/send-magic-link`)
    return response.data
  }

  // Message endpoints
  async getTicketMessages(ticketId: string, cursor?: string, limit?: number): Promise<MessagesResponse> {
    const params = new URLSearchParams()
    if (cursor) params.append('cursor', cursor)
    if (limit) params.append('limit', limit.toString())
    
    const response: AxiosResponse<MessagesResponse> = await this.client.get(
      `/tickets/${ticketId}/messages?${params.toString()}`
    )
    return response.data
  }

  async createMessage(ticketId: string, data: CreateMessageRequest): Promise<Message> {
    const response: AxiosResponse<Message> = await this.client.post(
      `/tickets/${ticketId}/messages`,
      {
        body: data.content,
        is_private: false
      }
    )
    return response.data
  }

  // Integration endpoints
  async getIntegrations(): Promise<Integration[]> {
    const response: AxiosResponse<Integration[]> = await this.client.get('/integrations')
    return response.data
  }

  async getIntegration(id: string): Promise<Integration> {
    const response: AxiosResponse<Integration> = await this.client.get(`/integrations/${id}`)
    return response.data
  }

  async createIntegration(data: Partial<Integration>): Promise<Integration> {
    const response: AxiosResponse<Integration> = await this.client.post('/integrations', data)
    return response.data
  }

  async updateIntegration(id: string, data: Partial<Integration>): Promise<Integration> {
    const response: AxiosResponse<Integration> = await this.client.put(`/integrations/${id}`, data)
    return response.data
  }

  async deleteIntegration(id: string): Promise<void> {
    await this.client.delete(`/integrations/${id}`)
  }

  // Enhanced integration endpoints
  async getIntegrationCategories(featured?: boolean): Promise<{ categories: IntegrationCategoryWithTemplates[] }> {
    const params = featured ? { featured: 'true' } : {}
    const response = await this.client.get('/integrations/categories', { params })
    return response.data
  }

  async getIntegrationTemplates(categoryId?: string, featured?: boolean): Promise<{ templates: IntegrationTemplate[] }> {
    const params: Record<string, string> = {}
    if (categoryId) params.category_id = categoryId
    if (featured !== undefined) params.featured = featured.toString()
    const response = await this.client.get('/integrations/templates', { params })
    return response.data
  }

  async getIntegrationTemplate(type: string): Promise<IntegrationTemplate> {
    const response: AxiosResponse<IntegrationTemplate> = await this.client.get(`/integrations/templates/${type}`)
    return response.data
  }

  async getIntegrationsWithTemplates(type?: string, status?: string): Promise<{ integrations: IntegrationWithTemplate[] }> {
    const params: Record<string, string> = {}
    if (type) params.type = type
    if (status) params.status = status
    const response = await this.client.get('/integrations/with-templates', { params })
    return response.data
  }

  async startOAuthFlow(integrationType: string, redirectUrl?: string): Promise<{ oauth_url: string; state: string }> {
    const response = await this.client.post('/integrations/oauth/start', {
      integration_type: integrationType,
      redirect_url: redirectUrl
    })
    return response.data
  }

  async handleOAuthCallback(type: string, code: string, state: string): Promise<Integration> {
    const response: AxiosResponse<Integration> = await this.client.post(`/integrations/${type}/oauth/callback`, {
      code,
      state
    })
    return response.data
  }

  async testIntegration(id: string): Promise<{ result: string; message: string }> {
    const response = await this.client.post(`/integrations/${id}/test`)
    return response.data
  }

  async getIntegrationMetrics(id: string): Promise<any> {
    const response = await this.client.get(`/integrations/${id}/metrics`)
    return response.data
  }

  // API Key endpoints (project-scoped)
  async getApiKeys(): Promise<ApiKey[]> {
    const response = await this.client.get('/api-keys')
    return response.data || []
  }

  async createApiKey(data: { name: string }): Promise<ApiKey & { key: string }> {
    const response = await this.client.post('/api-keys', data)
    return response.data
  }

  async updateApiKey(id: string, data: Partial<ApiKey>): Promise<ApiKey> {
    const response = await this.client.patch(`/api-keys/${id}`, data)
    return response.data.api_key
  }

  async deleteApiKey(id: string): Promise<void> {
    await this.client.delete(`/api-keys/${id}`)
  }

  // Email Inbox endpoints  
  async getEmailInbox(filter: EmailFilter = {}): Promise<EmailInboxResponse> {
    const params = new URLSearchParams()
    if (filter.search) params.append('search', filter.search)
    if (filter.mailbox) params.append('mailbox', filter.mailbox)
    if (filter.is_read !== undefined) params.append('is_read', filter.is_read.toString())
    if (filter.is_reply !== undefined) params.append('is_reply', filter.is_reply.toString())
    if (filter.has_attachments !== undefined) params.append('has_attachments', filter.has_attachments.toString())
    if (filter.from_date) params.append('from_date', filter.from_date)
    if (filter.to_date) params.append('to_date', filter.to_date)
    if (filter.thread_id) params.append('thread_id', filter.thread_id)
    if (filter.page) params.append('page', filter.page.toString())
    if (filter.limit) params.append('limit', filter.limit.toString())

    const response: AxiosResponse<EmailInboxResponse> = await this.client.get(`/email/inbox?${params}`)
    return response.data
  }

  async getEmailFromId(emailId: string): Promise<EmailInbox> {
    const response: AxiosResponse<EmailInbox> = await this.client.get(`/email/inbox/${emailId}`)
    return response.data
  }

  async syncEmails(): Promise<void> {
    await this.client.post('/email/inbox/sync')
  }

  async convertEmailToTicket(emailId: string, ticketData: ConvertToTicketRequest): Promise<void> {
    await this.client.post(`/email/inbox/${emailId}/convert-to-ticket`, ticketData)
  }

  async markEmailAsRead(emailId: string): Promise<void> {
    await this.client.post(`/email/inbox/${emailId}/mark-read`)
  }

  async replyToEmail(emailId: string, replyData: {
    body: string;
    subject?: string;
    cc_addresses?: string[];
    is_private?: boolean;
  }): Promise<void> {
    await this.client.post(`/email/inbox/${emailId}/reply`, replyData)
  }

  // Analytics endpoints
  async getAnalytics(period: string = '7d') {
    const response = await this.client.get(`/analytics?period=${period}`)
    return response.data
  }

  // Settings endpoints
  async getEmailSettings(): Promise<EmailSettings> {
    const response = await this.client.get('/settings/email')
    return response.data
  }

  async updateEmailSettings(data: EmailSettings): Promise<EmailSettings> {
    const response = await this.client.put('/settings/email', data)
    return response.data
  }

  async getBrandingSettings(): Promise<BrandingSettings> {
    const response = await this.client.get('/settings/branding')
    return response.data
  }

  async updateBrandingSettings(data: BrandingSettings): Promise<BrandingSettings> {
    const response = await this.client.put('/settings/branding', data)
    return response.data
  }

  async getAutomationSettings(): Promise<AutomationSettings> {
    const response = await this.client.get('/settings/automation')
    return response.data
  }

  async updateAutomationSettings(data: AutomationSettings): Promise<AutomationSettings> {
    const response = await this.client.put('/settings/automation', data)
    return response.data
  }

  async getAboutMeSettings(): Promise<AboutMeSettings> {
    const response = await this.client.get('/settings/about-me')
    return response.data
  }

  async updateAboutMeSettings(data: AboutMeSettings): Promise<AboutMeSettings> {
    const response = await this.client.put('/settings/about-me', data)
    return response.data
  }

  // Email Connector endpoints
  async getEmailConnectors(): Promise<{ connectors: EmailConnector[] }> {
    const response: AxiosResponse<{ connectors: EmailConnector[] }> = await this.client.get('/email/connectors')
    return response.data
  }

  async getEmailConnector(id: string): Promise<EmailConnector> {
    const response: AxiosResponse<EmailConnector> = await this.client.get(`/email/connectors/${id}`)
    return response.data
  }

  async createEmailConnector(data: EmailConnectorRequest): Promise<EmailConnector> {
    const response: AxiosResponse<EmailConnector> = await this.client.post('/email/connectors', data)
    return response.data
  }

  async updateEmailConnector(id: string, data: EmailConnectorRequest): Promise<EmailConnector> {
    const response: AxiosResponse<EmailConnector> = await this.client.patch(`/email/connectors/${id}`, data)
    return response.data
  }

  async deleteEmailConnector(id: string): Promise<void> {
    await this.client.delete(`/email/connectors/${id}`)
  }

  async validateEmailConnector(id: string, data: ValidateConnectorRequest): Promise<{ message: string }> {
    const response: AxiosResponse<{ message: string }> = await this.client.post(`/email/connectors/${id}/validate`, data)
    return response.data
  }

  async verifyEmailConnectorOTP(id: string, data: VerifyOTPRequest): Promise<{ message: string }> {
    const response: AxiosResponse<{ message: string }> = await this.client.post(`/email/connectors/${id}/verify-otp`, data)
    return response.data
  }

  // Email Mailbox endpoints
  async getEmailMailboxes(): Promise<{ mailboxes: EmailMailbox[] }> {
    const response: AxiosResponse<{ mailboxes: EmailMailbox[] }> = await this.client.get('/email/mailboxes')
    return response.data
  }

  async getEmailMailbox(id: string): Promise<EmailMailbox> {
    const response: AxiosResponse<EmailMailbox> = await this.client.get(`/email/mailboxes/${id}`)
    return response.data
  }

  async createEmailMailbox(data: EmailMailboxRequest): Promise<EmailMailbox> {
    const response: AxiosResponse<EmailMailbox> = await this.client.post('/email/mailboxes', data)
    return response.data
  }

  async updateEmailMailbox(id: string, data: EmailMailboxRequest): Promise<EmailMailbox> {
    const response: AxiosResponse<EmailMailbox> = await this.client.put(`/email/mailboxes/${id}`, data)
    return response.data
  }

  async deleteEmailMailbox(id: string): Promise<void> {
    await this.client.delete(`/email/mailboxes/${id}`)
  }

  // Domain Validation endpoints
  async getDomainValidations(): Promise<DomainValidation[]> {
    const response: AxiosResponse<{ domains: DomainValidation[] }> = await this.client.get(`/email/domains`)
    return response.data.domains
  }

  async createDomainValidation(data: { domain: string; }): Promise<DomainValidation> {
    const response: AxiosResponse<DomainValidation> = await this.client.post(`/email/domains`, data)
    return response.data
  }

  async verifyDomainValidation(domainId: string, data: { proof: string }): Promise<{ success: boolean; message: string }> {
    const response: AxiosResponse<{ success: boolean; message: string }> = await this.client.post(`/email/domains/${domainId}/verify`, data)
    return response.data
  }

  async deleteDomainValidation(domainId: string): Promise<void> {
    await this.client.delete(`/email/domains/${domainId}`)
  }

  async getClientChatStatus(sessionId: string): Promise<{status: string}> {
    const response: AxiosResponse<{status: string}> = await this.client.get(`/chat/sessions/${sessionId}/client/status`)
    return response.data
  }

  // Chat Widget endpoints
  async createChatWidget(data: CreateChatWidgetRequest): Promise<ChatWidget> {
    const response: AxiosResponse<ChatWidget> = await this.client.post('/chat/widgets', data)
    return response.data
  }

  async getChatWidget(widgetId: string): Promise<ChatWidget> {
    const response: AxiosResponse<ChatWidget> = await this.client.get(`/chat/widgets/${widgetId}`)
    return response.data
  }

  async listChatWidgets(): Promise<ChatWidget[]> {
    const response: AxiosResponse<{ widgets: ChatWidget[] }> = await this.client.get('/chat/widgets')
    return response.data.widgets
  }

  async updateChatWidget(widgetId: string, data: UpdateChatWidgetRequest): Promise<ChatWidget> {
    console.log('Updating widget:', widgetId, JSON.stringify(data))
    const response: AxiosResponse<ChatWidget> = await this.client.patch(`/chat/widgets/${widgetId}`, data)
    return response.data
  }

  async deleteChatWidget(widgetId: string): Promise<void> {
    await this.client.delete(`/chat/widgets/${widgetId}`)
  }

  // Chat Session endpoints
  async listChatSessions(filters?: ChatSessionFilters): Promise<ChatSession[]> {
    const params = new URLSearchParams()
    if (filters?.status) params.append('status', filters.status)
    if (filters?.assigned_agent_id) params.append('assigned_agent_id', filters.assigned_agent_id)
    if (filters?.widget_id) params.append('widget_id', filters.widget_id)
    if (filters?.limit) params.append('limit', filters.limit.toString())

    const response: AxiosResponse<{ sessions: ChatSession[] }> = await this.client.get(`/chat/sessions?${params.toString()}`)
    return response.data.sessions
  }

  async createChatSession(data: { widget_id: string; customer_email: string; customer_name?: string }): Promise<ChatSession> {
    const response: AxiosResponse<ChatSession> = await this.client.post('/chat/sessions', data)
    return response.data
  }

  async getChatSession(sessionId: string): Promise<ChatSession> {
    const response: AxiosResponse<ChatSession> = await this.client.get(`/chat/sessions/${sessionId}`)
    return response.data
  }

  async getActiveChatSessions(): Promise<ChatSession[]> {
    const response: AxiosResponse<{ sessions: ChatSession[] }> = await this.client.get('/chat/sessions/active')
    return response.data.sessions
  }

  async assignChatSession(sessionId: string, data: AssignChatSessionRequest): Promise<void> {
    await this.client.post(`/chat/sessions/${sessionId}/assign`, data)
  }

  async endChatSession(sessionId: string): Promise<void> {
    await this.client.post(`/chat/sessions/${sessionId}/end`)
  }

  // Chat Message endpoints
  async getChatMessages(sessionId: string, includePrivate: boolean = true): Promise<ChatMessage[]> {
    const params = includePrivate ? '?include_private=true' : ''
    const response: AxiosResponse<{ messages: ChatMessage[] }> = await this.client.get(`/chat/sessions/${sessionId}/messages${params}`)
    return response.data.messages
  }

  async sendChatMessage(sessionId: string, data: SendChatMessageRequest): Promise<ChatMessage> {
    const response: AxiosResponse<ChatMessage> = await this.client.post(`/chat/sessions/${sessionId}/messages`, data)
    return response.data
  }

  async markChatMessagesAsRead(sessionId: string, messageId: string): Promise<void> {
    await this.client.post(`/chat/sessions/${sessionId}/messages/${messageId}/read`, )
  }

  // Handoff methods
  async acceptHandoff(sessionId: string): Promise<HandoffResponse> {
    const response: AxiosResponse<HandoffResponse> = await this.client.post(`/chat/handoff/${sessionId}/accept`)
    return response.data
  }

  async declineHandoff(sessionId: string): Promise<HandoffResponse> {
    const response: AxiosResponse<HandoffResponse> = await this.client.post(`/chat/handoff/${sessionId}/decline`)
    return response.data
  }

  async getHandoffStatus(sessionId: string): Promise<HandoffStatus> {
    const response: AxiosResponse<HandoffStatus> = await this.client.get(`/chat/handoff/${sessionId}/status`)
    return response.data
  }

  // Notification methods
  async getNotifications(limit = 20, offset = 0): Promise<{ notifications: Notification[], limit: number, offset: number }> {
    const response: AxiosResponse<{ notifications: Notification[], limit: number, offset: number }> = 
      await this.client.get(`/notifications?limit=${limit}&offset=${offset}`)
    return response.data
  }

  async getNotificationCount(): Promise<NotificationCount> {
    const response: AxiosResponse<NotificationCount> = await this.client.get('/notifications/count')
    return response.data
  }

  async markNotificationAsRead(notificationId: string): Promise<void> {
    await this.client.put(`/notifications/${notificationId}/read`)
  }

  async markAllNotificationsAsRead(): Promise<void> {
    await this.client.put('/notifications/mark-all-read')
  }

  async getAIMetrics(): Promise<AIMetrics> {
    const response: AxiosResponse<AIMetrics> = await this.client.get('/chat/ai/metrics')
    return response.data
  }

  // Knowledge Management endpoints
  async uploadDocument(_projectId: string, file: File): Promise<KnowledgeDocument> {
    const formData = new FormData()
    formData.append('file', file)
    
    const response: AxiosResponse<KnowledgeDocument> = await this.client.post(
      `/knowledge/documents`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return response.data
  }

  async getDocuments(_projectId: string): Promise<KnowledgeDocument[]> {
    const response: AxiosResponse<{ documents: KnowledgeDocument[] }> = await this.client.get(
      `/knowledge/documents`
    )
    return response.data.documents
  }

  async deleteDocument(_projectId: string, documentId: string): Promise<void> {
    await this.client.delete(`/knowledge/documents/${documentId}`)
  }

  async createScrapingJob(_projectId: string, data: CreateScrapingJobRequest): Promise<KnowledgeScrapingJob> {
    const response: AxiosResponse<KnowledgeScrapingJob> = await this.client.post(
      `/knowledge/scrape`,
      data
    )
    return response.data
  }

  getScrapingJobStreamUrl(): string {
    const tenantId = localStorage.getItem('tenant_id') || this.tenantId
    const projectId = localStorage.getItem('project_id') || this.projectId
    if (!tenantId || !projectId) {
      throw new Error('Tenant and project must be selected before streaming scraping progress')
    }
    const baseUrl = this.client.defaults.baseURL || API_BASE_URL
    return `${baseUrl}/tenants/${tenantId}/projects/${projectId}/knowledge/scrape/stream`
  }

  async getScrapingJobs(_projectId: string): Promise<KnowledgeScrapingJob[]> {
    const response: AxiosResponse<{ jobs: KnowledgeScrapingJob[] }> = await this.client.get(
      `/knowledge/scraping-jobs`
    )
    return response.data.jobs
  }

  async getScrapingJobLinks(jobId: string): Promise<ScrapingJobLinksResponse> {
    const response: AxiosResponse<{ links: ScrapedLinkPreview[]; max_selectable_links?: number }> = await this.client.get(
      `/knowledge/scraping-jobs/${jobId}/links`
    )
    return {
      links: response.data.links || [],
      maxSelectableLinks: response.data.max_selectable_links ?? 10,
    }
  }

  async selectScrapingJobLinks(jobId: string, urls: string[]): Promise<SelectScrapingLinksResponse> {
    const response: AxiosResponse<{ selected_count: number; message: string; max_selectable_links?: number }> = await this.client.post(
      `/knowledge/scraping-jobs/${jobId}/select-links`,
      { urls }
    )
    return {
      selected_count: response.data.selected_count,
      message: response.data.message,
      maxSelectableLinks: response.data.max_selectable_links ?? 10,
    }
  }

  getScrapingJobIndexStreamUrl(jobId: string): string {
    const tenantId = localStorage.getItem('tenant_id') || this.tenantId
    const projectId = localStorage.getItem('project_id') || this.projectId

    if (!tenantId || !projectId) {
      throw new Error('Tenant and project must be selected before streaming indexing progress')
    }

    const baseUrl = this.client.defaults.baseURL || API_BASE_URL
    return `${baseUrl}/tenants/${tenantId}/projects/${projectId}/knowledge/scraping-jobs/${jobId}/index/stream`
  }

  async searchKnowledge(_projectId: string, query: string): Promise<KnowledgeSearchResult> {
    const response: AxiosResponse<KnowledgeSearchResult> = await this.client.get(
      `/knowledge/search?q=${encodeURIComponent(query)}`
    )
    return response.data
  }

  // WebSocket URL for real-time chat (agent endpoint)
  getChatWebSocketUrl(): string {
    const tenantId = localStorage.getItem('tenant_id')
    const projectId = localStorage.getItem('project_id')
    
    if (!tenantId || !projectId) {
      throw new Error('Tenant ID and Project ID are required for WebSocket connection')
    }
    
    const wsUrl = this.client.defaults.baseURL?.replace('http', 'ws') || 'ws://localhost:8080/v1'
    return `${wsUrl}/tenants/${tenantId}/projects/${projectId}/chat/ws`
  }

  // Alarm methods
  async getActiveAlarms(projectId: string): Promise<HowlingAlarm[]> {
    const tenantId = localStorage.getItem('tenant_id')
    if (!tenantId) {
      throw new Error('No tenant information available')
    }
    
    const response: AxiosResponse<{ alarms: HowlingAlarm[] }> = await this.client.get(
      `/tenants/${tenantId}/projects/${projectId}/alarms/active`
    )
    return response.data.alarms
  }

  async getAlarmStats(projectId: string): Promise<AlarmStats> {
    const tenantId = localStorage.getItem('tenant_id')
    if (!tenantId) {
      throw new Error('No tenant information available')
    }
    
    const response: AxiosResponse<AlarmStats> = await this.client.get(
      `/tenants/${tenantId}/projects/${projectId}/alarms/stats`
    )
    return response.data
  }

  async acknowledgeAlarm(projectId: string, alarmId: string, response?: string): Promise<void> {
    const tenantId = localStorage.getItem('tenant_id')
    if (!tenantId) {
      throw new Error('No tenant information available')
    }
    
    await this.client.post(
      `/tenants/${tenantId}/projects/${projectId}/alarms/${alarmId}/acknowledge`, 
      { response }
    )
  }

  // Notification Settings methods
  async getNotificationSettings(agentId: string): Promise<any> {
    const response: AxiosResponse<any> = await this.client.get(`/agents/${agentId}/notification-preferences`)
    return response.data
  }

  async updateNotificationSettings(agentId: string, settings: any): Promise<any> {
    const response: AxiosResponse<any> = await this.client.put(`/agents/${agentId}/notification-preferences`, settings)
    return response.data
  }

  // AI Theme Generation methods
  async scrapeWebsiteTheme(url: string): Promise<Partial<CreateChatWidgetRequest>> {
    const response: AxiosResponse<Partial<CreateChatWidgetRequest>> = await this.client.get(
      `/chat/widgets/scrape-theme?url=${encodeURIComponent(url)}`
    )
    return response.data
  }
}

export const apiClient = new APIClient()
export default apiClient
