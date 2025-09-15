import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Copy, CheckCircle, Code } from 'lucide-react'
import { useChatWidgetForm } from '../hooks/useChatWidgetForm'
import { PageHeader } from '../components/widget-form/PageHeader'
// import { BasicInformationSection } from '../components/widget-form/BasicInformationSection'
import { AgentPersonalizationSection } from '../components/widget-form/AgentPersonalizationSection'
import { FeaturesSection } from '../components/widget-form/FeaturesSection'
import { AppearanceSection } from '../components/widget-form/AppearanceSection'
import { WidgetSimulation } from '../components/widget-form/WidgetSimulation'
import { FormActions } from '../components/widget-form/FormActions'

export function CreateChatWidgetPage() {
  const navigate = useNavigate()
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [copiedCode, setCopiedCode] = useState(false)
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
      {domains.length > 0 && (
        <form onSubmit={handleSubmit}>
          {/* Form Grid Layout */}
          <div className="grid grid-cols-1 gap-8 xl:grid-cols-12">
            
            {/* Left Column - Form Sections */}
            <div className="xl:col-span-7 space-y-6">
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

              <FeaturesSection
                formData={formData}
                onUpdate={updateFormData}
              />

              <AppearanceSection
                formData={formData}
                onUpdate={updateFormData}
              />

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
                          Copy this code and paste it into your website's HTML
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
                        <div className="relative">
                          <div className="rounded-md border border-border bg-muted/50 p-4">
                            <pre className="text-sm font-mono leading-relaxed text-foreground overflow-x-auto whitespace-pre-wrap break-all">
                              {formData.embed_code}
                            </pre>
                          </div>
                          
                          {/* Copy button */}
                          <button
                            type="button"
                            onClick={copyEmbedCode}
                            className="absolute top-3 right-3 inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3"
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

                        {/* Instructions */}
                        <div className="space-y-2 text-sm text-muted-foreground">
                          <p className="font-medium">Integration Instructions:</p>
                          <ol className="list-decimal list-inside space-y-1 ml-2">
                            <li>Copy the embed code above</li>
                            <li>Paste it before the closing &lt;/body&gt; tag in your HTML</li>
                            <li>The chat widget will automatically appear on your website</li>
                            <li>Test the widget functionality after implementation</li>
                          </ol>
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
      )}
    </div>
  )
}