import React from 'react';
import { Helmet } from 'react-helmet-async';
import { HeroSection } from '../components/HeroSection';
import { FeaturesSection } from '../components/FeaturesSection';
import { PricingSection } from '../components/PricingSection';
import { TestimonialsSection } from '../components/TestimonialsSection';
import { CTASection } from '../components/CTASection';
import { Footer } from '../components/Footer';

export default function IndexPage() {
  return (
    <>
      <Helmet>
        <title>Hith | Modern Ticket Management & AI Chat</title>
        <meta
          name="description"
          content="Streamline your customer support with our integrated ticket management system and AI-powered chat platform. Fast, efficient, and intelligent solutions."
        />
      </Helmet>
      
      <main>
        <HeroSection />
        <FeaturesSection />
        <PricingSection />
        <TestimonialsSection />
        <CTASection />
      </main>
      
      <Footer />
    </>
  );
}