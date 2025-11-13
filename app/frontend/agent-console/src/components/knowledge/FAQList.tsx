import { KnowledgeFAQItem } from '../../lib/api'

interface FAQListProps {
  faqItems: KnowledgeFAQItem[]
}

export function FAQList({ faqItems }: FAQListProps) {
  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="space-y-4">
        <div>
          <h3 className="font-medium text-foreground">Knowledge Q&A</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Top questions and answers generated automatically from your website content
          </p>
        </div>

        <ul className="space-y-3">
          {faqItems.map((faq) => (
            <li key={faq.id} className="rounded-lg border border-border/60 bg-muted/20 p-4">
              <div className="text-sm font-semibold text-foreground">{faq.question}</div>
              <p className="mt-2 text-sm text-muted-foreground whitespace-pre-wrap">{faq.answer}</p>
              <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                {faq.source_url && (
                  <a
                    href={faq.source_url}
                    target="_blank"
                    rel="noreferrer"
                    className="text-primary hover:underline break-all"
                  >
                    {faq.source_url}
                  </a>
                )}
                {faq.source_section && (
                  <span className="rounded-full border border-border/60 bg-background px-2 py-0.5">
                    {faq.source_section}
                  </span>
                )}
                {faq.metadata?.category && (
                  <span className="rounded-full border border-border/60 bg-background px-2 py-0.5 capitalize">
                    {String(faq.metadata.category)}
                  </span>
                )}
                <span className="ml-auto text-xs text-muted-foreground">
                  {new Date(faq.created_at).toLocaleString()}
                </span>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
