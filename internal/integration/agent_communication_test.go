package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gaskaj/DeveloperAndQAAgent/internal/agent"
	"github.com/gaskaj/DeveloperAndQAAgent/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAgentCommunication(t *testing.T) {
	tests := []struct {
		name          string
		setupIssue    func(*TestEnvironment) *MockIssue
		expectedState state.WorkflowState
		timeout       time.Duration
	}{
		{
			name: "successful_issue_processing",
			setupIssue: func(te *TestEnvironment) *MockIssue {
				return te.SimulateGitHubIssue(123, "Test issue", "Simple test issue body", []string{"agent:ready"})
			},
			expectedState: state.StateComplete,
			timeout:       30 * time.Second,
		},
		{
			name: "complex_issue_decomposition",
			setupIssue: func(te *TestEnvironment) *MockIssue {
				te.claudeClient.SetResponse("", "COMPLEXITY_ASSESSMENT: TOO_COMPLEX\nThis requires decomposition.")
				return te.SimulateGitHubIssue(124, "Complex feature", "Very complex issue requiring multiple changes", []string{"agent:ready"})
			},
			expectedState: state.StateDecompose,
			timeout:       30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTestEnvironment(t)
			defer te.Cleanup()

			// Setup the issue
			_ = tt.setupIssue(te)

			// Create developer agent
			devAgent, err := te.CreateDeveloperAgent()
			require.NoError(t, err)

			// Create orchestrator
			orchestrator := te.CreateOrchestrator([]agent.Agent{devAgent})

			// Run agent processing with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Start orchestrator in background
			done := make(chan error, 1)
			go func() {
				done <- orchestrator.Run(ctx)
			}()

			// Wait for expected workflow state or timeout
			err = te.WaitForWorkflowTransition("developer", tt.expectedState, tt.timeout-5*time.Second)
			require.NoError(t, err)

			// Cancel context to stop orchestrator
			cancel()

			// Wait for orchestrator to finish
			select {
			case err := <-done:
				// Context cancellation is expected
				assert.True(t, errors.Is(err, context.Canceled) || err == nil)
			case <-time.After(5 * time.Second):
				t.Fatal("orchestrator did not stop gracefully")
			}

			// Verify final state
			te.AssertWorkflowState("developer", tt.expectedState)

			// For now, just verify basic agent functionality
			// Full workflow tests would require Claude integration
		})
	}
}

func TestAgentMessageProtocolCompliance(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		validate    func(*TestEnvironment, int)
	}{
		{
			name:        "claim_message_format",
			messageType: "claim",
			validate: func(te *TestEnvironment, issueNum int) {
				te.AssertCommentCreated(issueNum, "🤖 Developer agent claiming this issue")
				te.AssertIssueLabels(issueNum, []string{"agent:claimed"})
			},
		},
		{
			name:        "analysis_message_format",
			messageType: "analysis",
			validate: func(te *TestEnvironment, issueNum int) {
				te.AssertCommentCreated(issueNum, "🤖 **Analysis complete**")
			},
		},
		{
			name:        "failure_message_format",
			messageType: "failure",
			validate: func(te *TestEnvironment, issueNum int) {
				te.AssertCommentCreated(issueNum, "🤖 Developer agent failed:")
				te.AssertIssueLabels(issueNum, []string{"agent:failed"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTestEnvironment(t)
			defer te.Cleanup()

			issueNum := 125
			_ = te.SimulateGitHubIssue(issueNum, "Protocol test", "Testing message protocols", []string{"agent:ready"})

			// Simulate failure for failure test
			if tt.messageType == "failure" {
				te.SimulateAgentFailure("github", errors.New("simulated GitHub error"))
			}

			devAgent, err := te.CreateDeveloperAgent()
			require.NoError(t, err)

			orchestrator := te.CreateOrchestrator([]agent.Agent{devAgent})

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- orchestrator.Run(ctx)
			}()

			// Wait for processing to begin
			time.Sleep(2 * time.Second)

			// Clear any simulated errors after initial processing
			if tt.messageType == "failure" {
				te.githubClient.ClearError()
			}

			cancel()
			<-done

			// Validate protocol compliance
			tt.validate(te, issueNum)
		})
	}
}

func TestConcurrentAgentCommunication(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.Cleanup()

	// Create multiple agents (simulating multiple developer instances)
	agents := make([]agent.Agent, 3)
	for i := 0; i < 3; i++ {
		agent, err := te.CreateDeveloperAgent()
		require.NoError(t, err)
		agents[i] = agent
	}

	// Create multiple issues
	issues := make([]*MockIssue, 5)
	for i := 0; i < 5; i++ {
		issueNum := 200 + i
		issues[i] = te.SimulateGitHubIssue(issueNum, fmt.Sprintf("Concurrent issue %d", i),
			fmt.Sprintf("Test issue %d body", i), []string{"agent:ready"})
	}

	orchestrator := te.CreateOrchestrator(agents)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- orchestrator.Run(ctx)
	}()

	// Wait for all issues to be processed
	time.Sleep(30 * time.Second)
	cancel()
	<-done

	// Verify that issues were processed without conflicts
	processedCount := 0
	for _, issue := range issues {
		labels := te.githubClient.GetIssueLabels(issue.Number)
		labelMap := make(map[string]bool)
		for _, label := range labels {
			labelMap[label] = true
		}

		// Check if issue was claimed (minimum processing)
		if labelMap["agent:claimed"] {
			processedCount++
		}
	}

	// At least some issues should have been processed
	assert.Greater(t, processedCount, 0, "No issues were processed in concurrent scenario")
}

func TestAgentTimeoutHandling(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.Cleanup()

	// Create an agent with very long processing time
	te.claudeClient.SetMaxIterations(100) // Force long processing

	_ = te.SimulateGitHubIssue(130, "Timeout test", "This will timeout", []string{"agent:ready"})

	devAgent, err := te.CreateDeveloperAgent()
	require.NoError(t, err)

	orchestrator := te.CreateOrchestrator([]agent.Agent{devAgent})

	// Use a short timeout to force timeout scenario
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err = orchestrator.Run(ctx)
	duration := time.Since(start)

	// Should timeout within reasonable time
	assert.True(t, duration < 10*time.Second, "Agent did not respect timeout")
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))

	// Verify that the issue was at least claimed
	te.AssertIssueLabels(130, []string{"agent:claimed"})
}

func TestAgentRetryMechanisms(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.Cleanup()

	_ = te.SimulateGitHubIssue(140, "Retry test", "Testing retry mechanisms", []string{"agent:ready"})

	// Simulate temporary GitHub error
	simulatedErr := errors.New("temporary GitHub API error")
	te.SimulateAgentFailure("github", simulatedErr)

	devAgent, err := te.CreateDeveloperAgent()
	require.NoError(t, err)

	orchestrator := te.CreateOrchestrator([]agent.Agent{devAgent})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- orchestrator.Run(ctx)
	}()

	// Wait a bit, then clear the error to allow retry success
	time.Sleep(5 * time.Second)
	te.githubClient.ClearError()

	// Wait for completion or timeout
	time.Sleep(10 * time.Second)
	cancel()
	<-done

	// Verify that despite the initial error, the issue was eventually processed
	// (This depends on the error handling and retry logic in the actual implementation)
	comments := te.githubClient.GetIssueComments(140)
	assert.Greater(t, len(comments), 0, "No comments were created, suggesting retry mechanism failed")
}

func TestAgentStateConsistency(t *testing.T) {
	te := NewTestEnvironment(t)
	defer te.Cleanup()

	_ = te.SimulateGitHubIssue(150, "State consistency test", "Testing state consistency", []string{"agent:ready"})

	devAgent, err := te.CreateDeveloperAgent()
	require.NoError(t, err)

	orchestrator := te.CreateOrchestrator([]agent.Agent{devAgent})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- orchestrator.Run(ctx)
	}()

	// Monitor state transitions
	states := make([]state.WorkflowState, 0)
	stateMutex := sync.Mutex{}

	// Poll for state changes
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ws, err := te.store.Load(context.Background(), "developer")
				if err == nil && ws != nil {
					stateMutex.Lock()
					if len(states) == 0 || states[len(states)-1] != ws.State {
						states = append(states, ws.State)
					}
					stateMutex.Unlock()
				}
			}
		}
	}()

	// Wait for processing
	time.Sleep(20 * time.Second)
	cancel()
	<-done

	// Verify state transitions follow expected sequence
	stateMutex.Lock()
	defer stateMutex.Unlock()

	assert.Greater(t, len(states), 0, "No state transitions recorded")

	// Verify that states follow a logical progression
	expectedProgression := map[state.WorkflowState][]state.WorkflowState{
		state.StateClaim:     {state.StateWorkspace},
		state.StateWorkspace: {state.StateAnalyze},
		state.StateAnalyze:   {state.StateImplement, state.StateDecompose},
		state.StateImplement: {state.StateCommit, state.StateFailed},
		state.StateCommit:    {state.StatePR},
		state.StatePR:        {state.StateReview},
		state.StateReview:    {state.StateComplete},
	}

	for i := 0; i < len(states)-1; i++ {
		currentState := states[i]
		nextState := states[i+1]

		if validNext, exists := expectedProgression[currentState]; exists {
			found := false
			for _, valid := range validNext {
				if nextState == valid {
					found = true
					break
				}
			}
			assert.True(t, found,
				fmt.Sprintf("Invalid state transition from %s to %s", currentState, nextState))
		}
	}
}
