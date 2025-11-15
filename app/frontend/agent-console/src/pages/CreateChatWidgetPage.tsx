import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Copy, CheckCircle } from 'lucide-react'
import { CreateChatWidgetRequest, useChatWidgetForm } from '../hooks/useChatWidgetForm'
import { PageHeader } from '../components/widget-form/PageHeader'
// import { BasicInformationSection } from '../components/widget-form/BasicInformationSection'
import { AgentPersonalizationSection, isAgentPersonalizationComplete } from '../components/widget-form/AgentPersonalizationSection'
import { AppearanceSection, isAppearanceSectionComplete } from '../components/widget-form/AppearanceSection'
import { WidgetSimulation } from '../components/widget-form/WidgetSimulation'
import { FormActions } from '../components/widget-form/FormActions'
import { BuilderModeToggle, type BuilderMode } from '../components/widget-form/BuilderModeToggle'
import { AIBuilderSection } from '../components/widget-form/AIBuilderSection'

export function CreateChatWidgetPage() {
  const navigate = useNavigate()
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [copiedCode, setCopiedCode] = useState(false)
  const [builderMode, setBuilderMode] = useState<BuilderMode>('manual')
  const [aiLoading, setAiLoading] = useState(false)
  const [aiError, setAiError] = useState<string | null>(null)
  const [isAgentPersonalizationCollapsed, setIsAgentPersonalizationCollapsed] = useState(false)
  const [isAppearanceCollapsed, setIsAppearanceCollapsed] = useState(false)
  const {
    widgetId,
    domains,
    loading,
    submitting,
    error,
    formData,
    updateFormData,
    submitForm,
    setWidgetIdDynamic
  } = useChatWidgetForm()


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    const success = await submitForm()
    if (success) {
      navigate(`/chat/widgets?${widgetId ? 'updated' : 'created'}=true`)
    }
  }

  const handleCancel = () => {
    navigate('/chat/widgets')
  }

  const handleThemeGenerated = (theme: any) => {
    // Called when user clicks "Continue to Widget Settings" after AI build completes
    // Extract widget_id if present and update it dynamically
    if (theme.widget_id) {
      setWidgetIdDynamic(theme.widget_id)

      // Update URL without navigation to include widget ID
      window.history.replaceState(
        null,
        '',
        `/chat/widget/edit/${theme.widget_id}`
      )
    }

    // Update form data with all widget configuration
    const { widget_id, ...widgetData } = theme
    updateFormData(widgetData)

    setBuilderMode('manual') // Switch to manual mode to show the generated values
  }


  const handleModeChange = (mode: BuilderMode) => {
    setBuilderMode(mode)
    setAiError(null)
  }

  const toggleAgentPersonalizationCollapse = () => {
    setIsAgentPersonalizationCollapsed(!isAgentPersonalizationCollapsed)
  }

  const toggleAppearanceCollapse = () => {
    setIsAppearanceCollapsed(!isAppearanceCollapsed)
  }

  const copyEmbedCode = () => {
    if (!formData.embed_code) return
    
    navigator.clipboard.writeText(formData.embed_code)
    setCopiedCode(true)
    setSuccessMessage('Embed code copied to clipboard')
    setTimeout(() => {
      setCopiedCode(false)
      setSuccessMessage(null)
    }, 3000)
  }

  if (loading) {
    return (
      <div className="flex min-h-screen w-full items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent"></div>
          <p className="text-sm text-muted-foreground">Loading widget configuration...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full max-h-screen p-6 overflow-y-auto">

      {/* Alerts and Messages */}
      <PageHeader 
        widgetId={widgetId}
        error={error}
        domains={domains}
      />

      {/* Form Content */}
      
      <form onSubmit={handleSubmit}>
        {/* Builder Mode Toggle */}
        <BuilderModeToggle
          mode={builderMode}
          onChange={handleModeChange}
        />

        {/* Form Grid Layout */}
        <div className="grid grid-cols-1 gap-8 xl:grid-cols-12">

          {/* Left Column - Form Sections */}
          <div className="xl:col-span-5 space-y-6">
            {/* AI Builder Section */}
            {builderMode === 'ai' && (
              <AIBuilderSection
                onThemeGenerated={handleThemeGenerated}
                loading={aiLoading}
                error={aiError}
                onLoadingChange={setAiLoading}
                onError={setAiError}
                initialUrl={formData.domain_url}
              />
            )}

            {/* Manual Mode Sections */}
            {builderMode === 'manual' && (
              <>
                {/* <BasicInformationSection
                  formData={formData}
                  domains={domains}
                  widgetId={widgetId}
                  onUpdate={updateFormData}
                /> */}

                <AgentPersonalizationSection
                  formData={formData}
                  onUpdate={updateFormData}
                  isCollapsed={isAgentPersonalizationCollapsed}
                  onToggleCollapse={toggleAgentPersonalizationCollapse}
                />

                <AppearanceSection
                  formData={formData}
                  onUpdate={updateFormData}
                  isCollapsed={isAppearanceCollapsed}
                  onToggleCollapse={toggleAppearanceCollapse}
                />
              </>
            )}

            {/* Embedded Code Section */}
            {formData.embed_code && (
              <div className="rounded-lg border border-border bg-card shadow-sm">
                {/* Section Header */}
                <div className="border-b border-border bg-muted/50 px-6 py-4">
                  <div className="flex items-center gap-2">
                    <Copy className="h-5 w-5 text-primary" />
                    <h3 className="text-base font-semibold text-foreground">Embed Code</h3>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">Add this code to your website to enable the chat widget</p>
                </div>

                {/* Section Content */}
                <div className="p-6 space-y-4">
                  {/* Success message */}
                  {successMessage && (
                    <div className="flex items-center gap-2 p-3 rounded-md bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200">
                      <CheckCircle className="h-4 w-4" />
                      <span className="text-sm font-medium">{successMessage}</span>
                    </div>
                  )}

                  {/* Code block */}
                  <div className="rounded-md border border-border bg-muted/50 p-4">
                    <code className="text-sm font-mono text-foreground break-all">
                      {formData.embed_code}
                    </code>
                  </div>

                  {/* Instructions and Copy button */}
                  <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
                    <p className="text-sm text-muted-foreground">
                      Paste the code before &lt;/body&gt; tag in your HTML
                    </p>
                    <button
                      type="button"
                      onClick={copyEmbedCode}
                      className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4"
                      disabled={!formData.embed_code}
                    >
                      {copiedCode ? (
                        <>
                          <CheckCircle className="h-4 w-4 text-green-600" />
                          <span className="text-green-600">Copied!</span>
                        </>
                      ) : (
                        <>
                          <Copy className="h-4 w-4" />
                          <span>Copy Code</span>
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Form Actions - Only in Manual Mode */}
            {builderMode === 'manual' && (
              <FormActions
                submitting={submitting}
                widgetId={widgetId}
                onCancel={handleCancel}
              />
            )}
          </div>

          {/* Right Column - Live Preview */}
          <div className="xl:col-span-7">
            <div className="xl:sticky xl:top-0">
              <WidgetSimulation
                formData={formData}
                domains={domains}
              />
            </div>
          </div>
        </div>
      </form>
      
    </div>
  )
}