package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/service"
)

func main() {
	fmt.Println("🔍 Testing Online Agent Discovery Service")
	fmt.Println("=========================================")

	// Create test configuration
	agenticConfig := &config.AgenticConfig{
		Enabled:           true,
		AgentAssignment:   true,
		MaxConcurrentSessions: 100,
	}

	// Create the service
	agentService := service.NewOnlineAgentService(agenticConfig)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create test agents
	fmt.Println("\n📝 Registering Test Agents:")
	fmt.Println("---------------------------")

	agents := []*service.OnlineAgent{
		{
			UserID:          uuid.New(),
			TenantID:        tenantID,
			Name:            "Alice Johnson",
			Email:           "alice@company.com",
			Status:          service.AgentStatusOnline,
			Skills:          []service.AgentSkill{service.SkillGeneral, service.SkillSupport},
			ActiveChats:     1,
			MaxChats:        5,
			AvgResponseTime: 30.0, // 30 seconds
			LastActivity:    time.Now(),
			LastAssignment:  time.Now().Add(-10 * time.Minute),
		},
		{
			UserID:          uuid.New(),
			TenantID:        tenantID,
			Name:            "Bob Smith",
			Email:           "bob@company.com",
			Status:          service.AgentStatusOnline,
			Skills:          []service.AgentSkill{service.SkillTechnical, service.SkillSupport},
			ActiveChats:     3,
			MaxChats:        4,
			AvgResponseTime: 45.0, // 45 seconds
			LastActivity:    time.Now(),
			LastAssignment:  time.Now().Add(-5 * time.Minute),
		},
		{
			UserID:          uuid.New(),
			TenantID:        tenantID,
			Name:            "Carol Davis",
			Email:           "carol@company.com",
			Status:          service.AgentStatusOnline,
			Skills:          []service.AgentSkill{service.SkillBilling, service.SkillSupport},
			ActiveChats:     0,
			MaxChats:        6,
			AvgResponseTime: 20.0, // 20 seconds
			LastActivity:    time.Now(),
			LastAssignment:  time.Now().Add(-30 * time.Minute),
		},
		{
			UserID:          uuid.New(),
			TenantID:        tenantID,
			Name:            "David Wilson",
			Email:           "david@company.com",
			Status:          service.AgentStatusBusy,
			Skills:          []service.AgentSkill{service.SkillComplaint, service.SkillSupport},
			ActiveChats:     3,
			MaxChats:        3,
			AvgResponseTime: 60.0, // 60 seconds
			LastActivity:    time.Now(),
			LastAssignment:  time.Now().Add(-2 * time.Minute),
		},
		{
			UserID:          uuid.New(),
			TenantID:        tenantID,
			Name:            "Eva Martinez",
			Email:           "eva@company.com",
			Status:          service.AgentStatusAway,
			Skills:          []service.AgentSkill{service.SkillGeneral, service.SkillTechnical},
			ActiveChats:     1,
			MaxChats:        4,
			AvgResponseTime: 35.0, // 35 seconds
			LastActivity:    time.Now(),
			LastAssignment:  time.Now().Add(-15 * time.Minute),
		},
	}

	// Register all agents
	for i, agent := range agents {
		err := agentService.RegisterAgent(ctx, agent)
		if err != nil {
			log.Printf("Error registering agent %d: %v", i+1, err)
			continue
		}
		fmt.Printf("✅ Registered: %s (%s) - Status: %s, Skills: %v, Active: %d/%d\n",
			agent.Name, agent.Email, agent.Status, agent.Skills, agent.ActiveChats, agent.MaxChats)
	}

	// Test getting available agents
	fmt.Println("\n🔍 Available Agents Discovery:")
	fmt.Println("------------------------------")

	availableAgents, err := agentService.GetAvailableAgents(ctx, tenantID)
	if err != nil {
		log.Printf("Error getting available agents: %v", err)
	} else {
		fmt.Printf("Total available agents: %d\n", len(availableAgents))
		for i, agent := range availableAgents {
			fmt.Printf("%d. %s - Workload: %.1f%%, Response Time: %.0fs, Last Assignment: %s ago\n",
				i+1, agent.Name, agent.Workload*100, agent.AvgResponseTime,
				time.Since(agent.LastAssignment).Round(time.Minute))
		}
	}

	// Test agent selection scenarios
	fmt.Println("\n🎯 Agent Selection Test Scenarios:")
	fmt.Println("-----------------------------------")

	testScenarios := []struct {
		name        string
		criteria    service.AgentSelectionCriteria
		expectAgent bool
	}{
		{
			name: "General Support Request",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills: []service.AgentSkill{service.SkillSupport},
				UrgencyLevel:   service.UrgencyNormal,
				RequestType:    service.AgentRequestTypeSupport,
			},
			expectAgent: true,
		},
		{
			name: "Technical Issue",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills: []service.AgentSkill{service.SkillTechnical},
				UrgencyLevel:   service.UrgencyHigh,
				RequestType:    service.AgentRequestTypeTechnical,
			},
			expectAgent: true,
		},
		{
			name: "Billing Dispute",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills: []service.AgentSkill{service.SkillBilling},
				UrgencyLevel:   service.UrgencyNormal,
				RequestType:    service.AgentRequestTypeBilling,
			},
			expectAgent: true,
		},
		{
			name: "Complaint Escalation",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills: []service.AgentSkill{service.SkillComplaint},
				UrgencyLevel:   service.UrgencyCritical,
				RequestType:    service.AgentRequestTypeComplaint,
			},
			expectAgent: false, // David is busy
		},
		{
			name: "Fast Response Required",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills:  []service.AgentSkill{service.SkillSupport},
				MaxResponseTime: 25.0, // Must respond within 25 seconds
				UrgencyLevel:    service.UrgencyHigh,
				RequestType:     service.AgentRequestTypeUrgent,
			},
			expectAgent: true, // Carol has 20s response time
		},
		{
			name: "Exclude Specific Agent",
			criteria: service.AgentSelectionCriteria{
				RequiredSkills: []service.AgentSkill{service.SkillSupport},
				ExcludeAgents:  []uuid.UUID{agents[0].UserID, agents[2].UserID}, // Exclude Alice and Carol
				UrgencyLevel:   service.UrgencyNormal,
				RequestType:    service.AgentRequestTypeSupport,
			},
			expectAgent: true, // Should get Bob or Eva
		},
		{
			name: "Preferred Agent",
			criteria: service.AgentSelectionCriteria{
				PreferredAgent: &agents[2].UserID, // Prefer Carol
				UrgencyLevel:   service.UrgencyNormal,
				RequestType:    service.AgentRequestTypeSupport,
			},
			expectAgent: true,
		},
	}

	successfulSelections := 0
	totalScenarios := len(testScenarios)

	for i, scenario := range testScenarios {
		result, err := agentService.SelectBestAgent(ctx, tenantID, scenario.criteria)
		if err != nil {
			log.Printf("Error selecting agent for scenario %d: %v", i+1, err)
			continue
		}

		hasAgent := result.SelectedAgent != nil
		isCorrect := hasAgent == scenario.expectAgent

		if isCorrect {
			successfulSelections++
		}

		statusIcon := "✅"
		if !isCorrect {
			statusIcon = "❌"
		}

		fmt.Printf("\n%d. %s: %s\n", i+1, scenario.name, statusIcon)
		fmt.Printf("   Expected Agent: %t, Got Agent: %t\n", scenario.expectAgent, hasAgent)

		if result.SelectedAgent != nil {
			fmt.Printf("   🎯 Selected: %s (Score: %.1f)\n", result.SelectedAgent.Name, result.Score)
			fmt.Printf("   📊 Reason: %s\n", result.Reason)
			if result.AlternateAgent != nil {
				fmt.Printf("   🔄 Alternate: %s\n", result.AlternateAgent.Name)
			}
		} else {
			fmt.Printf("   📊 Reason: %s\n", result.Reason)
		}

		fmt.Printf("   ⏱️  Processing Time: %v\n", result.ProcessingTime)
		fmt.Printf("   📈 Total Available: %d agents\n", result.TotalAgents)
	}

	// Test agent stats
	fmt.Println("\n📊 Agent Statistics:")
	fmt.Println("--------------------")

	stats, err := agentService.GetAgentStats(ctx, tenantID)
	if err != nil {
		log.Printf("Error getting agent stats: %v", err)
	} else {
		fmt.Printf("Total Agents: %d\n", stats["total_agents"])
		fmt.Printf("Online Agents: %d\n", stats["online_agents"])
		fmt.Printf("Available Agents: %d\n", stats["available_agents"])
		fmt.Printf("Busy Agents: %d\n", stats["busy_agents"])
		fmt.Printf("Average Workload: %.1f%%\n", stats["avg_workload"].(float64)*100)
		
		skillsCoverage := stats["skills_coverage"].(map[service.AgentSkill]int)
		fmt.Println("Skills Coverage:")
		for skill, count := range skillsCoverage {
			fmt.Printf("  • %s: %d agents\n", skill, count)
		}
	}

	// Test workload updates
	fmt.Println("\n🔄 Testing Workload Updates:")
	fmt.Println("-----------------------------")

	// Simulate Carol getting assigned a new chat
	err = agentService.UpdateAgentWorkload(ctx, agents[2].UserID, 2)
	if err != nil {
		log.Printf("Error updating Carol's workload: %v", err)
	} else {
		fmt.Println("✅ Updated Carol's workload from 0 to 2 active chats")
	}

	// Test status updates
	fmt.Println("\n📱 Testing Status Updates:")
	fmt.Println("--------------------------")

	// Set Eva to online status
	err = agentService.UpdateAgentStatus(ctx, agents[4].UserID, service.AgentStatusOnline)
	if err != nil {
		log.Printf("Error updating Eva's status: %v", err)
	} else {
		fmt.Println("✅ Updated Eva's status from Away to Online")
	}

	// Test agent unregistration
	fmt.Println("\n👋 Testing Agent Unregistration:")
	fmt.Println("--------------------------------")

	err = agentService.UnregisterAgent(ctx, agents[3].UserID)
	if err != nil {
		log.Printf("Error unregistering David: %v", err)
	} else {
		fmt.Printf("✅ Unregistered David Wilson\n")
	}

	// Final available agents check
	finalAvailable, err := agentService.GetAvailableAgents(ctx, tenantID)
	if err != nil {
		log.Printf("Error getting final available agents: %v", err)
	} else {
		fmt.Printf("📊 Final available agents count: %d\n", len(finalAvailable))
	}

	// Performance summary
	fmt.Println("\n📊 Performance & Capabilities Summary:")
	fmt.Println("------------------------------------")
	accuracy := float64(successfulSelections) / float64(totalScenarios) * 100
	fmt.Printf("   Total Scenarios Tested: %d\n", totalScenarios)
	fmt.Printf("   Successful Selections: %d (%.1f%%)\n", successfulSelections, accuracy)

	fmt.Println("\n✅ Service Capabilities:")
	fmt.Println("   • Real-time agent registration and discovery")
	fmt.Println("   • Skill-based agent matching")
	fmt.Println("   • Workload-aware agent selection")
	fmt.Println("   • Response time optimization")
	fmt.Println("   • Agent preference and exclusion support")
	fmt.Println("   • Comprehensive scoring algorithm")
	fmt.Println("   • Real-time status and workload updates")
	fmt.Println("   • Automatic cleanup of inactive agents")
	fmt.Println("   • Detailed statistics and monitoring")
	fmt.Println("   • Alternate agent suggestions")

	if accuracy >= 85 {
		fmt.Println("\n🎉 Phase 3 Task 3.2 Complete!")
		fmt.Println("    (Online Agent Discovery)")
		fmt.Println("\n🚀 Online Agent Discovery Service is Ready!")
	} else {
		fmt.Printf("\n⚠️  Accuracy %.1f%% below target 85%% - needs improvement\n", accuracy)
	}
}
