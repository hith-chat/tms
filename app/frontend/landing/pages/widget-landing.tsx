import React from 'react';
import { Helmet } from 'react-helmet-async';
import { WidgetHero } from '../components/WidgetHero';
import { WidgetFeatures } from '../components/WidgetFeatures';
import { WidgetShowcase } from '../components/WidgetShowcase';
import { CTASection } from '../components/CTASection';
import { Footer } from '../components/Footer';

export default function WidgetLandingPage() {
  return (
    <>
      <Helmet>
        <title>AI Chat Widget for Websites | Convert Visitors to Customers</title>
        <meta
          name="description"
          content="AI-powered chat widget that answers questions, schedules meetings, and generates leads automatically. Set up in 60 seconds. Free forever plan available."
        />
        <meta
          name="keywords"
          content="AI chat widget, customer support, lead generation, chatbot, website chat"
        />
      </Helmet>

      <main>
        <WidgetHero />
        <WidgetFeatures />
        <WidgetShowcase />
        <CTASection />
      </main>

      <Footer />
    </>
  );
}
