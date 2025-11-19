import { useState, useEffect } from 'react'
import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { 
  Zap, 
  Plus, 
  Search, 
  Settings, 
  ExternalLink,
  Check,
  AlertCircle,
  Star,
  Power,
  Trash2,
  RotateCw,
  Eye,
  MessageCircle,
  Folder,
  Users,
  Headphones,
  Calendar,
  Cloud,
  CreditCard,
  Mail,
  Code,
  ShoppingCart,
  BarChart,
  Share2,
} from 'lucide-react'
import { 
  apiClient, 
  IntegrationCategoryWithTemplates, 
  IntegrationTemplate, 
  IntegrationWithTemplate
} from '../lib/api'

// Tab types for integration navigation
type IntegrationsTab = 'browse' | 'installed' | 'marketplace'

interface IntegrationCardProps {
  integration: IntegrationWithTemplate
  onConfigure: (integration: IntegrationWithTemplate) => void
  onToggle: (integration: IntegrationWithTemplate) => void
  onDelete: (integration: IntegrationWithTemplate) => void
  onTest: (integration: IntegrationWithTemplate) => void
}

interface TemplateCardProps {
  template: IntegrationTemplate
  onInstall: (template: IntegrationTemplate) => void
}

function IntegrationCard({ integration, onConfigure, onToggle, onDelete, onTest }: IntegrationCardProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-green-700 bg-green-100 border-green-200'
      case 'inactive': return 'text-gray-700 bg-gray-100 border-gray-200'
      case 'error': return 'text-red-700 bg-red-100 border-red-200'
      case 'configuring': return 'text-yellow-700 bg-yellow-100 border-yellow-200'
      default: return 'text-gray-700 bg-gray-100 border-gray-200'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <Check className="w-4 h-4" />
      case 'inactive': return <Power className="w-4 h-4" />
      case 'error': return <AlertCircle className="w-4 h-4" />
      case 'configuring': return <RotateCw className="w-4 h-4 animate-spin" />
      default: return <AlertCircle className="w-4 h-4" />
    }
  }

  return (
    <div className="bg-background border border-border rounded-xl p-6 hover:shadow-lg hover:border-primary/20 transition-all duration-200 group">
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-4 flex-1">
          {integration.template?.logo_url ? (
            <div className="relative">
              <img 
                src={integration.template.logo_url} 
                alt={integration.name}
                className="w-14 h-14 rounded-xl object-contain border border-border group-hover:border-primary/20 transition-colors"
              />
              <div className={`absolute -bottom-1 -right-1 w-5 h-5 rounded-full border-2 border-background flex items-center justify-center ${getStatusColor(integration.status)}`}>
                {getStatusIcon(integration.status)}
              </div>
            </div>
          ) : (
            <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-primary/10 to-primary/20 flex items-center justify-center border border-primary/20">
              <Zap className="w-7 h-7 text-primary" />
            </div>
          )}
          
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between mb-2">
              <div>
                <h3 className="text-lg font-semibold text-fg group-hover:text-primary transition-colors">
                  {integration.name}
                </h3>
                <div className="flex items-center space-x-2 mt-1">
                  <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(integration.status)}`}>
                    {getStatusIcon(integration.status)}
                    <span className="ml-1.5 capitalize">{integration.status}</span>
                  </span>
                  <span className="text-xs text-fg-muted bg-muted px-2 py-1 rounded-full">
                    {integration.type}
                  </span>
                </div>
              </div>
            </div>
            
            <p className="text-sm text-fg-muted mb-3 line-clamp-2">
              {integration.template?.description || `${integration.type} integration for seamless connectivity`}
            </p>
            
            {integration.last_error && (
              <div className="mb-3 p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-700 flex items-start">
                  <AlertCircle className="w-4 h-4 mr-2 mt-0.5 flex-shrink-0" />
                  {integration.last_error}
                </p>
              </div>
            )}
            
            <div className="flex items-center justify-between text-xs text-fg-muted">
              <div className="flex items-center space-x-4">
                <span className="flex items-center">
                  <span className="w-2 h-2 bg-green-500 rounded-full mr-1.5"></span>
                  Created {new Date(integration.created_at).toLocaleDateString()}
                </span>
                {integration.last_sync_at && (
                  <span className="flex items-center">
                    <RotateCw className="w-3 h-3 mr-1.5" />
                    Synced {new Date(integration.last_sync_at).toLocaleDateString()}
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
        
        {/* Action buttons */}
        <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => onTest(integration)}
            className="p-2.5 text-fg-muted hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
            title="Test connection"
          >
            <Eye className="w-4 h-4" />
          </button>
          <button
            onClick={() => onConfigure(integration)}
            className="p-2.5 text-fg-muted hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
            title="Configure"
          >
            <Settings className="w-4 h-4" />
          </button>
          <button
            onClick={() => onToggle(integration)}
            className={`p-2.5 rounded-lg transition-colors ${
              integration.status === 'active'
                ? 'text-fg-muted hover:text-orange-600 hover:bg-orange-50'
                : 'text-fg-muted hover:text-green-600 hover:bg-green-50'
            }`}
            title={integration.status === 'active' ? 'Disable' : 'Enable'}
          >
            <Power className="w-4 h-4" />
          </button>
          <button
            onClick={() => onDelete(integration)}
            className="p-2.5 text-fg-muted hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            title="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

function TemplateCard({ template, onInstall }: TemplateCardProps) {
  const getAuthMethodBadge = (method: string) => {
    switch (method) {
      case 'oauth':
        return (
          <span className="inline-flex items-center px-2.5 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-full">
            <Check className="w-3 h-3 mr-1" />
            OAuth
          </span>
        )
      case 'api_key':
        return (
          <span className="inline-flex items-center px-2.5 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded-full">
            <Settings className="w-3 h-3 mr-1" />
            API Key
          </span>
        )
      case 'none':
        return (
          <span className="inline-flex items-center px-2.5 py-1 bg-gray-100 text-gray-700 text-xs font-medium rounded-full">
            No Auth
          </span>
        )
      default:
        return null
    }
  }

  return (
    <div className="bg-background border border-border rounded-xl p-6 hover:shadow-lg hover:border-primary/20 transition-all duration-200 group cursor-pointer">
      <div className="flex items-start justify-between h-full">
        <div className="flex items-start space-x-4 flex-1">
          {template.logo_url ? (
            <div className="relative">
              <img 
                src={template.logo_url} 
                alt={template.display_name}
                className="w-14 h-14 rounded-xl object-contain border border-border group-hover:border-primary/20 transition-colors"
              />
              {template.is_featured && (
                <div className="absolute -top-1 -right-1 w-5 h-5 bg-yellow-500 rounded-full flex items-center justify-center">
                  <Star className="w-3 h-3 text-white" fill="currentColor" />
                </div>
              )}
            </div>
          ) : (
            <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-primary/10 to-primary/20 flex items-center justify-center border border-primary/20 relative">
              <Zap className="w-7 h-7 text-primary" />
              {template.is_featured && (
                <div className="absolute -top-1 -right-1 w-5 h-5 bg-yellow-500 rounded-full flex items-center justify-center">
                  <Star className="w-3 h-3 text-white" fill="currentColor" />
                </div>
              )}
            </div>
          )}
          
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between mb-2">
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-fg group-hover:text-primary transition-colors">
                  {template.display_name}
                </h3>
                <div className="flex items-center space-x-2 mt-1">
                  {getAuthMethodBadge(template.auth_method)}
                  {template.supported_events && template.supported_events.length > 0 && (
                    <span className="text-xs text-fg-muted bg-muted px-2 py-1 rounded-full">
                      {template.supported_events.length} events
                    </span>
                  )}
                </div>
              </div>
            </div>
            
            <p className="text-sm text-fg-muted mb-4 line-clamp-2">
              {template.description || 'No description available'}
            </p>
            
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                {template.website_url && (
                  <a
                    href={template.website_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-primary hover:text-primary/80 flex items-center transition-colors"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <ExternalLink className="w-3 h-3 mr-1" />
                    Learn more
                  </a>
                )}
                {template.documentation_url && (
                  <a
                    href={template.documentation_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-fg-muted hover:text-fg flex items-center transition-colors"
                    onClick={(e) => e.stopPropagation()}
                  >
                    📖 Docs
                  </a>
                )}
              </div>
              
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onInstall(template)
                }}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors opacity-0 group-hover:opacity-100 flex items-center text-sm font-medium"
              >
                <Plus className="w-4 h-4 mr-1.5" />
                Install
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export function IntegrationsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [activeTab, setActiveTab] = useState<IntegrationsTab>((searchParams.get('tab') as IntegrationsTab) || 'browse')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')

  // Categories and templates state
  const [categories, setCategories] = useState<IntegrationCategoryWithTemplates[]>([])
  const [installedIntegrations, setInstalledIntegrations] = useState<IntegrationWithTemplate[]>([])

  // Modal states
  const [showInstallModal, setShowInstallModal] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<IntegrationTemplate | null>(null)
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationWithTemplate | null>(null)

  const tabs = [
    { id: 'browse' as IntegrationsTab, name: 'Browse', icon: Search },
    { id: 'installed' as IntegrationsTab, name: 'Installed', icon: Settings },
    { id: 'marketplace' as IntegrationsTab, name: 'Marketplace', icon: Zap },
  ]

  const icons: Record<string, React.ElementType> = {
    browse: Search,
    installed: Settings,
    marketplace: Zap,
    message: MessageCircle,
    folder: Folder,
    users: Users,
    headphones: Headphones,
    calendar: Calendar,
    cloud: Cloud,
    "credit-card": CreditCard,
    mail: Mail,
    code: Code,
    "shopping-cart": ShoppingCart,
    zap: Zap,
    'bar-chart': BarChart,
    'share-2': Share2,
    settings: Settings,
    'message-circle': MessageCircle,
    // slack: Slack,
    
  }

  const handleTabChange = (tabId: IntegrationsTab) => {
    setActiveTab(tabId)
    setSearchParams({ tab: tabId })
  }

  useEffect(() => {
    loadData()
  }, [activeTab])

  const loadData = async () => {
    setLoading(true)
    setError(null)
    
    try {
      if (activeTab === 'browse' || activeTab === 'marketplace') {
        // Load categories with templates
        const featured = activeTab === 'marketplace' ? true : undefined
        const response = await apiClient.getIntegrationCategories(featured)
        setCategories(response.categories)
      } else if (activeTab === 'installed') {
        // Load installed integrations
        const response = await apiClient.getIntegrationsWithTemplates()
        setInstalledIntegrations(response.integrations)
      }
    } catch (err) {
      setError('Failed to load integrations data')
      console.error('Integrations load error:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleInstallIntegration = async (template: IntegrationTemplate) => {
    // For OAuth integrations, get the OAuth URL and redirect
    if (template.auth_method === 'oauth') {
      try {
        setLoading(true)
        const response = await apiClient.getIntegrationInstallUrl(template.type)
        // Redirect to OAuth provider
        window.location.href = response.oauth_url
      } catch {
        setError(`Failed to start ${template.display_name} installation`)
        setLoading(false)
      }
      return
    }

    // For non-OAuth integrations, show the configuration modal
    setSelectedTemplate(template)
    setShowInstallModal(true)
  }

  const handleConfigureIntegration = (integration: IntegrationWithTemplate) => {
    setSelectedIntegration(integration)
    setShowConfigModal(true)
  }

  const handleToggleIntegration = async (integration: IntegrationWithTemplate) => {
    try {
      const newStatus = integration.status === 'active' ? 'inactive' : 'active'
      await apiClient.updateIntegration(integration.id, { status: newStatus })
      loadData() // Reload data
    } catch (err) {
      setError('Failed to toggle integration')
    }
  }

  const handleDeleteIntegration = async (integration: IntegrationWithTemplate) => {
    if (!confirm(`Are you sure you want to delete the ${integration.name} integration?`)) {
      return
    }

    try {
      await apiClient.deleteIntegration(integration.id)
      loadData() // Reload data
    } catch (err) {
      setError('Failed to delete integration')
    }
  }

  const handleTestIntegration = async (integration: IntegrationWithTemplate) => {
    try {
      const result = await apiClient.testIntegration(integration.id)
      alert(`Test result: ${result.message}`)
    } catch (err) {
      setError('Integration test failed')
    }
  }

  const startOAuthFlow = async (template: IntegrationTemplate) => {
    try {
      const response = await apiClient.startOAuthFlow(template.type)
      // Redirect to OAuth URL
      window.location.href = response.oauth_url
    } catch (err) {
      setError('Failed to start OAuth flow')
    }
  }

  const filteredCategories = categories.filter(category => {
    if (selectedCategory !== 'all' && category.id !== selectedCategory) {
      return false
    }
    
    if (!searchQuery) return true
    
    const searchLower = searchQuery.toLowerCase()
    return (
      category.display_name.toLowerCase().includes(searchLower) ||
      category.templates.some(template => 
        template.display_name.toLowerCase().includes(searchLower) ||
        template.description?.toLowerCase().includes(searchLower)
      )
    )
  })

  const filteredIntegrations = installedIntegrations.filter(integration => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return (
      integration.name.toLowerCase().includes(searchLower) ||
      integration.type.toLowerCase().includes(searchLower) ||
      integration.template?.display_name.toLowerCase().includes(searchLower)
    )
  })

  const renderBrowseTab = () => (
    <div className="space-y-8">
      {/* Search and filters */}
      <div className="sticky top-0 bg-background/95 backdrop-blur-sm z-10 pb-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-fg-muted w-4 h-4" />
            <input
              type="text"
              placeholder="Search integrations, categories, or features..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent bg-background"
            />
          </div>
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="px-4 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent bg-background min-w-[180px]"
          >
            <option value="all">All Categories</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.display_name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Results summary */}
      {searchQuery && (
        <div className="flex items-center justify-between py-2 border-b border-border">
          <p className="text-fg-muted">
            Found {filteredCategories.reduce((acc, cat) => acc + cat.templates.filter(t => {
              const searchLower = searchQuery.toLowerCase()
              return t.display_name.toLowerCase().includes(searchLower) || 
                     t.description?.toLowerCase().includes(searchLower)
            }).length, 0)} results for "{searchQuery}"
          </p>
          <button
            onClick={() => setSearchQuery('')}
            className="text-primary hover:text-primary/80 text-sm"
          >
            Clear search
          </button>
        </div>
      )}

      {/* Categories with templates */}
      {filteredCategories.length === 0 ? (
        <div className="text-center py-16">
          <Search className="w-16 h-16 text-fg-muted mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-medium text-fg mb-2">No integrations found</h3>
          <p className="text-fg-muted mb-4">Try adjusting your search terms or browse all categories</p>
          <button
            onClick={() => {
              setSearchQuery('')
              setSelectedCategory('all')
            }}
            className="px-4 py-2 text-primary hover:text-primary/80 transition-colors"
          >
            Show all integrations
          </button>
        </div>
      ) : (
        <div className="space-y-12">
          {filteredCategories.map((category) => {
            const filteredTemplates = category.templates.filter(template => {
              if (!searchQuery) return true
              const searchLower = searchQuery.toLowerCase()
              return (
                template.display_name.toLowerCase().includes(searchLower) ||
                template.description?.toLowerCase().includes(searchLower)
              )
            })
            

            if (filteredTemplates.length === 0) return null

            return (
              <div key={category.id} className="space-y-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <h2 className="text-2xl font-semibold text-fg flex items-center">
                      {category.icon && icons[category.icon] && (
                        <span className="mr-3 text-2xl">
                          {React.createElement(icons[category.icon])}
                        </span>
                      )}
                      {category.display_name}
                      <span className="ml-3 text-base font-normal text-fg-muted bg-muted px-3 py-1 rounded-full">
                        {filteredTemplates.length}
                      </span>
                    </h2>
                  </div>
                  {category.templates.length > 4 && (
                    <button className="text-primary hover:text-primary/80 text-sm font-medium flex items-center">
                      View all <ExternalLink className="w-3 h-3 ml-1" />
                    </button>
                  )}
                </div>
                
                {category.description && (
                  <p className="text-fg-muted text-lg max-w-3xl">{category.description}</p>
                )}

                <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
                  {filteredTemplates.map((template) => (
                    <TemplateCard
                      key={template.id}
                      template={template}
                      onInstall={handleInstallIntegration}
                    />
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )

  const renderInstalledTab = () => (
    <div className="space-y-6">
      {/* Search and filters */}
      <div className="sticky top-0 bg-background/95 backdrop-blur-sm z-10 pb-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-fg-muted w-4 h-4" />
            <input
              type="text"
              placeholder="Search installed integrations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent bg-background"
            />
          </div>
          <div className="flex gap-2">
            <select
              className="px-4 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent bg-background"
              defaultValue="all"
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="error">Error</option>
            </select>
          </div>
        </div>
      </div>

      {/* Summary stats */}
      {installedIntegrations.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="text-2xl font-bold text-green-600">
              {installedIntegrations.filter(i => i.status === 'active').length}
            </div>
            <div className="text-sm text-green-700">Active</div>
          </div>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <div className="text-2xl font-bold text-gray-600">
              {installedIntegrations.filter(i => i.status === 'inactive').length}
            </div>
            <div className="text-sm text-gray-700">Inactive</div>
          </div>
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="text-2xl font-bold text-red-600">
              {installedIntegrations.filter(i => i.status === 'error').length}
            </div>
            <div className="text-sm text-red-700">Error</div>
          </div>
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="text-2xl font-bold text-blue-600">
              {installedIntegrations.length}
            </div>
            <div className="text-sm text-blue-700">Total</div>
          </div>
        </div>
      )}

      {/* Installed integrations */}
      {filteredIntegrations.length === 0 ? (
        <div className="text-center py-16">
          {installedIntegrations.length === 0 ? (
            <>
              <Zap className="w-16 h-16 text-fg-muted mx-auto mb-4 opacity-50" />
              <h3 className="text-xl font-medium text-fg mb-2">No integrations installed</h3>
              <p className="text-fg-muted mb-6 max-w-md mx-auto">
                Get started by browsing our catalog of 100+ integrations to connect with your favorite tools and services.
              </p>
              <button
                onClick={() => handleTabChange('browse')}
                className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors flex items-center mx-auto"
              >
                <Plus className="w-4 h-4 mr-2" />
                Browse Integrations
              </button>
            </>
          ) : (
            <>
              <Search className="w-16 h-16 text-fg-muted mx-auto mb-4 opacity-50" />
              <h3 className="text-lg font-medium text-fg mb-2">No matching integrations</h3>
              <p className="text-fg-muted mb-4">Try adjusting your search terms</p>
              <button
                onClick={() => setSearchQuery('')}
                className="px-4 py-2 text-primary hover:text-primary/80 transition-colors"
              >
                Clear search
              </button>
            </>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {/* Results count */}
          <div className="flex items-center justify-between py-2">
            <p className="text-fg-muted">
              Showing {filteredIntegrations.length} of {installedIntegrations.length} integrations
            </p>
            <div className="flex items-center space-x-2 text-sm">
              <span className="text-fg-muted">Sort by:</span>
              <select className="border-0 bg-transparent text-fg focus:ring-0 cursor-pointer">
                <option>Name</option>
                <option>Status</option>
                <option>Date added</option>
                <option>Last sync</option>
              </select>
            </div>
          </div>

          {/* Integration cards */}
          <div className="space-y-3">
            {filteredIntegrations.map((integration) => (
              <IntegrationCard
                key={integration.id}
                integration={integration}
                onConfigure={handleConfigureIntegration}
                onToggle={handleToggleIntegration}
                onDelete={handleDeleteIntegration}
                onTest={handleTestIntegration}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )

  const renderMarketplaceTab = () => (
    <div className="space-y-8">
      {/* Hero section */}
      <div className="text-center py-12 bg-gradient-to-br from-primary/5 to-purple-500/5 rounded-2xl border border-primary/10">
        <div className="max-w-2xl mx-auto">
          <div className="flex justify-center mb-6">
            <div className="relative">
              <Star className="w-20 h-20 text-yellow-500" fill="currentColor" />
              <div className="absolute -top-2 -right-2 w-6 h-6 bg-primary rounded-full flex items-center justify-center">
                <Star className="w-3 h-3 text-white" fill="currentColor" />
              </div>
            </div>
          </div>
          <h2 className="text-3xl font-bold text-fg mb-4">Featured Integrations</h2>
          <p className="text-lg text-fg-muted mb-6">
            Discover the most popular and trusted integrations used by thousands of teams worldwide
          </p>
          <div className="flex items-center justify-center space-x-8 text-sm text-fg-muted">
            <div className="flex items-center">
              <Check className="w-4 h-4 text-green-500 mr-2" />
              Verified & Secure
            </div>
            <div className="flex items-center">
              <Zap className="w-4 h-4 text-primary mr-2" />
              Quick Setup
            </div>
            <div className="flex items-center">
              <Star className="w-4 h-4 text-yellow-500 mr-2" />
              Community Favorites
            </div>
          </div>
        </div>
      </div>

      {/* Featured integrations grid */}
      <div className="space-y-8">
        <div className="flex items-center justify-between">
          <h3 className="text-xl font-semibold text-fg">🏆 Most Popular</h3>
          <span className="text-sm text-fg-muted">Trusted by 10,000+ teams</span>
        </div>
        
        {/* This would show only featured templates */}
        {renderBrowseTab()}
      </div>

      {/* Call to action */}
      <div className="text-center py-12 border-t border-border">
        <h3 className="text-xl font-semibold text-fg mb-2">Looking for something specific?</h3>
        <p className="text-fg-muted mb-6">
          Browse our complete catalog of 100+ integrations or request a custom integration
        </p>
        <div className="flex items-center justify-center space-x-4">
          <button
            onClick={() => handleTabChange('browse')}
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors flex items-center"
          >
            <Search className="w-4 h-4 mr-2" />
            Browse All Integrations
          </button>
          <button className="px-6 py-3 border border-border text-fg hover:bg-muted/50 rounded-lg transition-colors flex items-center">
            <Plus className="w-4 h-4 mr-2" />
            Request Integration
          </button>
        </div>
      </div>
    </div>
  )

  return (
    <div className="h-full flex flex-col bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
      {/* Enhanced Header with gradient and glass effect */}
      <div className="border-b border-border/50 bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60 shadow-sm">
        <div className="px-6 py-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              <div className="relative">
                <div className="absolute -inset-1 bg-gradient-to-r from-indigo-600 to-cyan-600 rounded-lg blur opacity-25"></div>
                <div className="relative p-3 bg-gradient-to-br from-indigo-50 to-cyan-50 dark:from-indigo-950 dark:to-cyan-950 rounded-lg border border-indigo-200/50 dark:border-indigo-800/50">
                  <Zap className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
                </div>
              </div>
              <div>
                <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-cyan-600 bg-clip-text text-transparent">
                  Integrations
                </h1>
                <div className="flex items-center gap-3 mt-1">
                  <p className="text-sm text-muted-foreground">
                    Connect with 100+ tools and services to streamline your workflow
                  </p>
                </div>
              </div>
            </div>
            
            {/* Quick stats */}
            <div className="flex items-center space-x-6 text-sm">
              <div className="text-center">
                <div className="font-semibold text-lg text-foreground">{categories.reduce((acc, cat) => acc + cat.templates.length, 0)}</div>
                <div className="text-muted-foreground">Available</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-lg text-green-600 dark:text-green-400">{installedIntegrations.filter(i => i.status === 'active').length}</div>
                <div className="text-muted-foreground">Active</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-lg text-foreground">{installedIntegrations.length}</div>
                <div className="text-muted-foreground">Installed</div>
              </div>
            </div>
          </div>

          {/* Error message */}
          {error && (
            <div className="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center">
              <AlertCircle className="w-4 h-4 mr-2 flex-shrink-0" />
              <span>{error}</span>
              <button 
                onClick={() => setError(null)}
                className="ml-auto text-red-500 hover:text-red-700"
              >
                ×
              </button>
            </div>
          )}

          {/* Tab navigation */}
          <div className="mt-6 flex space-x-1 bg-muted/50 p-1 rounded-lg w-fit">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={`flex items-center px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'bg-background text-fg shadow-sm'
                    : 'text-fg-muted hover:text-fg hover:bg-background/50'
                }`}
              >
                <tab.icon className="w-4 h-4 mr-2" />
                {tab.name}
                {tab.id === 'installed' && installedIntegrations.length > 0 && (
                  <span className="ml-2 bg-primary text-primary-foreground text-xs px-2 py-0.5 rounded-full">
                    {installedIntegrations.length}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content - Scrollable */}
      <div className="flex-1 overflow-y-auto">
        <div className="max-w-7xl mx-auto p-6">
          {/* Tab content */}
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <div className="text-center">
                <RotateCw className="w-8 h-8 animate-spin text-primary mx-auto mb-4" />
                <p className="text-fg-muted">Loading integrations...</p>
              </div>
            </div>
          ) : (
            <div className="pb-6">
              {activeTab === 'browse' && renderBrowseTab()}
              {activeTab === 'installed' && renderInstalledTab()}
              {activeTab === 'marketplace' && renderMarketplaceTab()}
            </div>
          )}
        </div>
      </div>

      {/* Installation Modal */}
      {showInstallModal && selectedTemplate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-background rounded-xl shadow-2xl p-6 max-w-md w-full max-h-[90vh] overflow-y-auto">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center">
                {selectedTemplate.logo_url && (
                  <img 
                    src={selectedTemplate.logo_url} 
                    alt={selectedTemplate.display_name}
                    className="w-8 h-8 rounded mr-3"
                  />
                )}
                <h3 className="text-lg font-semibold text-fg">
                  Install {selectedTemplate.display_name}
                </h3>
              </div>
              <button
                onClick={() => setShowInstallModal(false)}
                className="text-fg-muted hover:text-fg p-1"
              >
                ×
              </button>
            </div>
            
            <p className="text-fg-muted mb-6">
              {selectedTemplate.description}
            </p>

            {selectedTemplate.website_url && (
              <div className="mb-4 p-3 bg-muted/30 rounded-lg">
                <a 
                  href={selectedTemplate.website_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-primary hover:text-primary/80 text-sm flex items-center"
                >
                  <ExternalLink className="w-4 h-4 mr-1" />
                  Learn more about {selectedTemplate.display_name}
                </a>
              </div>
            )}
            
            <div className="flex items-center justify-end space-x-3">
              <button
                onClick={() => setShowInstallModal(false)}
                className="px-4 py-2 text-fg-muted hover:text-fg transition-colors"
              >
                Cancel
              </button>
              {selectedTemplate.auth_method === 'oauth' ? (
                <button
                  onClick={() => startOAuthFlow(selectedTemplate)}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors flex items-center"
                >
                  <Zap className="w-4 h-4 mr-2" />
                  Connect with OAuth
                </button>
              ) : (
                <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors flex items-center">
                  <Settings className="w-4 h-4 mr-2" />
                  Configure Manually
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Configuration Modal */}
      {showConfigModal && selectedIntegration && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-background rounded-xl shadow-2xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center">
                {selectedIntegration.template?.logo_url && (
                  <img 
                    src={selectedIntegration.template.logo_url} 
                    alt={selectedIntegration.name}
                    className="w-8 h-8 rounded mr-3"
                  />
                )}
                <h3 className="text-lg font-semibold text-fg">
                  Configure {selectedIntegration.name}
                </h3>
              </div>
              <button
                onClick={() => setShowConfigModal(false)}
                className="text-fg-muted hover:text-fg p-1"
              >
                ×
              </button>
            </div>
            
            <div className="space-y-4">
              <div className="p-4 bg-muted/30 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-fg">Status</span>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    selectedIntegration.status === 'active' 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {selectedIntegration.status}
                  </span>
                </div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-fg">Type</span>
                  <span className="text-sm text-fg-muted">{selectedIntegration.type}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-fg">Created</span>
                  <span className="text-sm text-fg-muted">
                    {new Date(selectedIntegration.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
              
              <p className="text-fg-muted">
                Integration configuration options will be available here based on the integration type and requirements.
              </p>
            </div>
            
            <div className="flex items-center justify-end space-x-3 mt-6">
              <button
                onClick={() => setShowConfigModal(false)}
                className="px-4 py-2 text-fg-muted hover:text-fg transition-colors"
              >
                Close
              </button>
              <button 
                onClick={() => handleTestIntegration(selectedIntegration)}
                className="px-4 py-2 border border-border text-fg hover:bg-muted/50 rounded-lg transition-colors flex items-center"
              >
                <Eye className="w-4 h-4 mr-2" />
                Test Connection
              </button>
              <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors flex items-center">
                <Check className="w-4 h-4 mr-2" />
                Save Settings
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
