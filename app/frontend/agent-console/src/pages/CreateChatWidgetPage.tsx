import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Copy, CheckCircle, Code } from 'lucide-react'
import { CreateChatWidgetRequest, useChatWidgetForm } from '../hooks/useChatWidgetForm'
import { PageHeader } from '../components/widget-form/PageHeader'
// import { BasicInformationSection } from '../components/widget-form/BasicInformationSection'
import { AgentPersonalizationSection } from '../components/widget-form/AgentPersonalizationSection'
import { AppearanceSection } from '../components/widget-form/AppearanceSection'
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
  const {
    widgetId,
    domains,
    loading,
    submitting,
    error,
    formData,
    updateFormData,
    submitForm
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

  const handleThemeGenerated = (theme: Partial<CreateChatWidgetRequest>) => {
    updateFormData(theme)
    setBuilderMode('manual') // Switch to manual mode to show the generated values
  }


  const handleModeChange = (mode: BuilderMode) => {
    setBuilderMode(mode)
    setAiError(null)
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
          <div className="xl:col-span-7 space-y-6">
            {/* AI Builder Section */}
            {builderMode === 'ai' && (
              <AIBuilderSection
                onThemeGenerated={handleThemeGenerated}
                loading={aiLoading}
                error={aiError}
                onLoadingChange={setAiLoading}
                onError={setAiError}
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
                />

                <AppearanceSection
                  formData={formData}
                  onUpdate={updateFormData}
                />
              </>
            )}

            {/* Embedded Code Section */}
            {formData.embed_code && (
              <div className="flex flex-col w-full min-w-0">
                {/* Card container with enterprise styling */}
                <div className="rounded-lg border border-border bg-card text-card-foreground shadow-sm">
                  {/* Header */}
                  <div className="flex items-center gap-3 p-6 pb-4">
                    <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10">
                      <Code className="h-4 w-4 text-primary" aria-hidden="true" />
                    </div>
                    <div className="flex flex-col space-y-1">
                      <h3 className="text-base font-semibold leading-none tracking-tight">
                        Embed Code
                      </h3>
                      <p className="text-sm text-muted-foreground">
                        Copy this single line and paste it before the closing &lt;/body&gt; tag in your HTML
                      </p>
                    </div>
                  </div>

                  {/* Code content */}
                  <div className="px-6 pb-6">
                    <div className="space-y-4">
                      {/* Success message */}
                      {successMessage && (
                        <div className="flex items-center gap-2 p-3 rounded-md bg-green-50 border border-green-200 text-green-800">
                          <CheckCircle className="h-4 w-4" />
                          <span className="text-sm font-medium">{successMessage}</span>
                        </div>
                      )}

                      {/* Code block */}
                      <div className="flex items-center gap-3">
                        <div className="flex-1 rounded-md border border-border bg-muted/80 p-4">
                          <code className="text-sm font-mono text-foreground break-all">
                            {formData.embed_code}
                          </code>
                        </div>

                        {/* Copy button */}
                        <button
                          type="button"
                          onClick={copyEmbedCode}
                          className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4"
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
                </div>
              </div>
            )}
          </div>

          {/* Right Column - Live Preview */}
          <div className="xl:col-span-5">
            <div className="xl:sticky xl:top-0">
              <WidgetSimulation
                formData={formData}
                domains={domains}
              />
            </div>
          </div>
        </div>

        {/* Form Actions */}
        <div className="border-t border-border pt-6 mt-8">
          <FormActions
            submitting={submitting}
            widgetId={widgetId}
            onCancel={handleCancel}
          />
        </div>
      </form>
      
    </div>
  )
}