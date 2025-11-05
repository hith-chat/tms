import React, { useState } from "react";
import { Button } from "../components/Button";
import { Input } from "../components/Input";
import { Textarea } from "../components/Textarea";
import { Mail, MessageSquare, CheckCircle, AlertCircle } from "lucide-react";

type SubmitResponse = {
  success: boolean;
  message: string;
  ticket_id?: string;
  ticket_url?: string;
};

export default function ContactPage() {
  const [formData, setFormData] = useState({
    title: "",
    body: "",
    email: "",
  });
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<SubmitResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      const res = await fetch("https://api.bareuptime.co/support/submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Accept": "*/*",
        },
        body: JSON.stringify(formData),
      });

      const data: SubmitResponse = await res.json();
      
      if (res.ok && data.success) {
        setResponse(data);
        // Reset form on success
        setFormData({ title: "", body: "", email: "" });
      } else {
        setError(data.message || "Failed to submit ticket");
      }
    } catch (err: any) {
      console.error("Contact form error:", err);
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  return (
    <div style={{ padding: 20, display: 'flex', justifyContent: 'center' }}>
      <div style={{ maxWidth: 800, width: '100%' }}>
        {/* Header */}
        <div style={{ textAlign: 'center', marginBottom: 48 }}>
          <div style={{ 
            background: 'var(--primary)', 
            color: 'var(--primary-foreground)', 
            padding: 16, 
            borderRadius: 12,
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: 24
          }}>
            <Mail size={32} />
          </div>
          <h1 style={{ fontSize: 36, fontWeight: 800, margin: 0, marginBottom: 12 }}>
            Contact Support
          </h1>
          <p style={{ fontSize: 18, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.6 }}>
            Need help? Send us a message and we'll get back to you as soon as possible.
          </p>
        </div>

        {/* Success Response */}
        {response && (
          <div style={{ 
            background: 'color-mix(in srgb, #10b981 10%, transparent)', 
            border: '1px solid #10b981', 
            padding: 20, 
            borderRadius: 12, 
            marginBottom: 32,
            display: 'flex',
            alignItems: 'flex-start',
            gap: 12
          }}>
            <CheckCircle size={24} style={{ color: '#10b981', flexShrink: 0, marginTop: 2 }} />
            <div>
              <h3 style={{ fontSize: 18, fontWeight: 700, margin: 0, marginBottom: 8, color: '#10b981' }}>
                Ticket Created Successfully!
              </h3>
              <p style={{ fontSize: 14, margin: 0, marginBottom: 12, lineHeight: 1.5 }}>
                Your support ticket has been created with ID: <strong>{response.ticket_id}</strong>
              </p>
              {response.ticket_url && (
                <a 
                  href={response.ticket_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  style={{ 
                    color: '#10b981', 
                    textDecoration: 'none', 
                    fontWeight: 600,
                    fontSize: 14
                  }}
                >
                  View Ticket →
                </a>
              )}
            </div>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div style={{ 
            background: 'color-mix(in srgb, #ef4444 10%, transparent)', 
            border: '1px solid #ef4444', 
            padding: 20, 
            borderRadius: 12, 
            marginBottom: 32,
            display: 'flex',
            alignItems: 'center',
            gap: 12
          }}>
            <AlertCircle size={24} style={{ color: '#ef4444', flexShrink: 0 }} />
            <div>
              <h3 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4, color: '#ef4444' }}>
                Error
              </h3>
              <p style={{ fontSize: 14, margin: 0, lineHeight: 1.5 }}>
                {error}
              </p>
            </div>
          </div>
        )}

        {/* Contact Form */}
        <div style={{ 
          background: 'var(--surface)', 
          border: '1px solid var(--border)', 
          padding: 32, 
          borderRadius: 16,
          boxShadow: '0 10px 30px rgba(16,24,40,0.08)'
        }}>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            {/* Email Field */}
            <div>
              <label style={{ 
                display: 'block', 
                fontSize: 14, 
                fontWeight: 600, 
                marginBottom: 8,
                color: 'var(--foreground)'
              }}>
                Email Address *
              </label>
              <Input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                placeholder="your.email@example.com"
                required
                style={{ width: '100%' }}
              />
            </div>

            {/* Title Field */}
            <div>
              <label style={{ 
                display: 'block', 
                fontSize: 14, 
                fontWeight: 600, 
                marginBottom: 8,
                color: 'var(--foreground)'
              }}>
                Subject *
              </label>
              <Input
                type="text"
                name="title"
                value={formData.title}
                onChange={handleChange}
                placeholder="Brief description of your issue"
                required
                style={{ width: '100%' }}
              />
            </div>

            {/* Message Field */}
            <div>
              <label style={{ 
                display: 'block', 
                fontSize: 14, 
                fontWeight: 600, 
                marginBottom: 8,
                color: 'var(--foreground)'
              }}>
                Message *
              </label>
              <Textarea
                name="body"
                value={formData.body}
                onChange={handleChange}
                placeholder="Describe your issue in detail. Include any relevant information that might help us assist you better."
                required
                rows={6}
                style={{ width: '100%', resize: 'vertical' }}
              />
            </div>

            {/* Submit Button */}
            <Button 
              type="submit" 
              disabled={loading || !formData.title.trim() || !formData.body.trim() || !formData.email.trim()}
              style={{ 
                padding: '12px 32px', 
                fontSize: 16, 
                fontWeight: 700,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 8
              }}
            >
              <MessageSquare size={20} />
              {loading ? 'Submitting...' : 'Send Message'}
            </Button>
          </form>
        </div>

        {/* Help Text */}
        <div style={{ 
          textAlign: 'center', 
          marginTop: 32,
          padding: 24,
          background: 'color-mix(in srgb, var(--primary) 5%, transparent)',
          borderRadius: 12,
          border: '1px solid color-mix(in srgb, var(--primary) 20%, transparent)'
        }}>
          <h3 style={{ fontSize: 18, fontWeight: 700, margin: 0, marginBottom: 8 }}>
            Need Immediate Help?
          </h3>
          <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
            For urgent issues, you can also reach us directly at{' '}
            <a href="mailto:support@hith.co" style={{ color: 'var(--primary)', textDecoration: 'none', fontWeight: 600 }}>
              support@hith.chat
            </a>
            {' '}or check our{' '}
            <a href="/privacy" style={{ color: 'var(--primary)', textDecoration: 'none', fontWeight: 600 }}>
              FAQ section
            </a>
            {' '}for quick answers.
          </p>
        </div>
      </div>
    </div>
  );
}