import React from 'react';
import { Helmet } from 'react-helmet-async';
import { WidgetHero } from '../components/WidgetHero';
import { WidgetFeatures } from '../components/WidgetFeatures';
import { Footer } from '../components/Footer';

export default function IndexPage() {
  return (
    <>
      <Helmet>
        <title>AI Chat Widget for Websites | Hith</title>
        <meta
          name="description"
          content="AI-powered chat widget that answers questions, schedules meetings, and generates leads automatically. Set up in 60 seconds."
        />
      </Helmet>

      <main>
        <WidgetHero />
        <WidgetFeatures />
      </main>

      <Footer />
    </>
  );
}