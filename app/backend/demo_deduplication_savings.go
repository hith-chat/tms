package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== EMBEDDING REDUNDANCY ANALYSIS ===")
	fmt.Println()
	
	// Scenario: 5 websites scraped 30 times each
	websites := 5
	scrapeRuns := 30
	avgPagesPerSite := 10
	
	// Cost calculations
	costPerEmbedding := 0.0001 // $0.0001 per OpenAI embedding call
	
	// Without deduplication
	totalCallsWithoutDedup := websites * scrapeRuns * avgPagesPerSite
	costWithoutDedup := float64(totalCallsWithoutDedup) * costPerEmbedding
	
	// With deduplication
	totalCallsWithDedup := websites * avgPagesPerSite
	costWithDedup := float64(totalCallsWithDedup) * costPerEmbedding
	
	// Savings
	savings := costWithoutDedup - costWithDedup
	savingsPercent := (savings / costWithoutDedup) * 100
	
	fmt.Printf("🌐 Websites: %d\n", websites)
	fmt.Printf("🔄 Scrape runs per website: %d\n", scrapeRuns)
	fmt.Printf("📄 Average pages per website: %d\n", avgPagesPerSite)
	fmt.Println()
	
	fmt.Printf("❌ WITHOUT DEDUPLICATION:\n")
	fmt.Printf("   Total embedding calls: %d\n", totalCallsWithoutDedup)
	fmt.Printf("   Total cost: $%.4f\n", costWithoutDedup)
	fmt.Println()
	
	fmt.Printf("✅ WITH DEDUPLICATION:\n")
	fmt.Printf("   Total embedding calls: %d\n", totalCallsWithDedup)
	fmt.Printf("   Total cost: $%.4f\n", costWithDedup)
	fmt.Println()
	
	fmt.Printf("💰 SAVINGS:\n")
	fmt.Printf("   Cost saved: $%.4f\n", savings)
	fmt.Printf("   Percentage saved: %.1f%%\n", savingsPercent)
	fmt.Printf("   Redundant calls avoided: %d\n", totalCallsWithoutDedup - totalCallsWithDedup)
	fmt.Println()
	
	fmt.Println("=== DATABASE IMPACT ===")
	fmt.Printf("📊 Vector storage without dedup: %d embeddings\n", totalCallsWithoutDedup)
	fmt.Printf("📊 Vector storage with dedup: %d embeddings\n", totalCallsWithDedup)
	fmt.Printf("💾 Storage reduction: %.1fx smaller\n", float64(totalCallsWithoutDedup)/float64(totalCallsWithDedup))
	fmt.Println()
	
	fmt.Println("=== SOLUTION IMPLEMENTED ===")
	fmt.Println("✅ Added project-level URL deduplication")
	fmt.Println("✅ Check existing URLs before inserting new ones")
	fmt.Println("✅ Skip duplicate URLs to prevent redundant embeddings")
	fmt.Println("✅ Log duplicate count for monitoring")
}
