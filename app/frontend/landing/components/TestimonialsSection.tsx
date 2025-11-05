import React from 'react';
import { Star } from 'lucide-react';
import { Avatar, AvatarImage, AvatarFallback } from './Avatar';
import styles from './TestimonialsSection.module.css';

const testimonials = [
  {
    name: 'Sunil Agarwal',
    title: 'CTO, Penify',
    avatar: 'https://www.penify.dev/_next/static/media/sunil.dd2e3068.webp',
    fallback: 'SA',
    rating: 5,
    quote: 'This platform is extremely easy to integrate. I was able to set it up in minutes and start managing customer queries efficiently.',
  },
  {
    name: 'Akansha Sinha',
    title: 'CEO, BareUptime',
    avatar: 'https://www.penify.dev/_next/static/media/akansha.68cfed9b.webp',
    fallback: 'AK',
    rating: 5,
    quote: 'The AI chat is incredibly powerful. It handles most of our common request in my uptime monitoring, allowing me to focus on product development.',
  },
  {
    name: 'Emily Rodriguez',
    title: 'Founder, NBC Universal',
    avatar: 'https://randomuser.me/api/portraits/women/33.jpg',
    fallback: 'ER',
    rating: 5,
    quote: 'As a small team, efficiency is key. This tool has helped how we handle customer support, making it faster and more effective.',
  },
];

const StarRating = ({ rating }: { rating: number }) => (
  <div className={styles.starRating}>
    {Array.from({ length: 5 }, (_, i) => (
      <Star
        key={i}
        size={16}
        className={i < rating ? styles.filledStar : styles.emptyStar}
      />
    ))}
  </div>
);

export const TestimonialsSection = ({ className }: { className?: string }) => {
  return (
    <section className={`${styles.testimonials} ${className || ''}`}>
      <div className={styles.container}>
        <div className={styles.header}>
          <h2 className={styles.title}>Loved by teams worldwide</h2>
          <p className={styles.subtitle}>
            Don't just take our word for it. Here's what our customers are saying.
          </p>
        </div>
        <div className={styles.grid}>
          {testimonials.map((testimonial, index) => (
            <div key={index} className={styles.card}>
              <StarRating rating={testimonial.rating} />
              <blockquote className={styles.quote}>
                "{testimonial.quote}"
              </blockquote>
              <div className={styles.author}>
                <Avatar>
                  <AvatarImage src={testimonial.avatar} alt={testimonial.name} />
                  <AvatarFallback>{testimonial.fallback}</AvatarFallback>
                </Avatar>
                <div className={styles.authorInfo}>
                  <p className={styles.authorName}>{testimonial.name}</p>
                  <p className={styles.authorTitle}>{testimonial.title}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};