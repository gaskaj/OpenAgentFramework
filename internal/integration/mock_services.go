package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gaskaj/DeveloperAndQAAgent/internal/ghub"
	"github.com/gaskaj/DeveloperAndQAAgent/internal/state"
	"github.com/google/go-github/v68/github"
)

// MockGitHubClient implements ghub.Client interface for testing
type MockGitHubClient struct {
	mu               sync.RWMutex
	issues           map[int]*MockIssue
	comments         map[int][]*MockComment
	prs              []*MockPR
	simulatedError   error
	assignees        map[int][]string
}

// MockIssue represents a mock GitHub issue
type MockIssue struct {
	Number    int
	Title     string
	Body      string
	Labels    []string
	Assignees []string
	State     string
}

// MockComment represents a mock GitHub comment
type MockComment struct {
	Body      string
	CreatedAt time.Time
}

// MockPR represents a mock GitHub pull request
type MockPR struct {
	Number int
	Title  string
	Body   string
	Head   string
	Base   string
	State  string
}

func (mc *MockComment) Contains(substring string) bool {
	return strings.Contains(mc.Body, substring)
}

// NewMockGitHubClient creates a new mock GitHub client
func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		issues:   make(map[int]*MockIssue),
		comments: make(map[int][]*MockComment),
		assignees: make(map[int][]string),
	}
}

func (m *MockGitHubClient) AddIssue(issue *MockIssue) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.issues[issue.Number] = issue
}

func (m *MockGitHubClient) GetMockIssue(number int) *MockIssue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.issues[number]
}

func (m *MockGitHubClient) GetIssueLabels(number int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if issue, ok := m.issues[number]; ok {
		return issue.Labels
	}
	return nil
}

func (m *MockGitHubClient) GetIssueComments(number int) []*MockComment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.comments[number]
}

func (m *MockGitHubClient) GetCreatedPRs() []*MockPR {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.prs
}

func (m *MockGitHubClient) SimulateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulatedError = err
}

func (m *MockGitHubClient) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulatedError = nil
}

// Implement ghub.Client interface
// ListIssues is an alias for ListIssuesByLabels for interface compatibility
func (m *MockGitHubClient) ListIssues(ctx context.Context, labels []string) ([]*github.Issue, error) {
	return m.ListIssuesByLabels(ctx, labels)
}

func (m *MockGitHubClient) ListIssuesByLabels(ctx context.Context, labels []string) ([]*github.Issue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	var result []*github.Issue
	for _, issue := range m.issues {
		// Check if issue has all required labels
		hasAllLabels := true
		for _, requiredLabel := range labels {
			found := false
			for _, issueLabel := range issue.Labels {
				if issueLabel == requiredLabel {
					found = true
					break
				}
			}
			if !found {
				hasAllLabels = false
				break
			}
		}
		
		if hasAllLabels {
			// Convert MockIssue to github.Issue
			number := issue.Number
			title := issue.Title
			body := issue.Body
			
			var githubLabels []*github.Label
			for _, labelName := range issue.Labels {
				name := labelName
				githubLabels = append(githubLabels, &github.Label{Name: &name})
			}
			
			var githubAssignees []*github.User
			for _, assignee := range issue.Assignees {
				login := assignee
				githubAssignees = append(githubAssignees, &github.User{Login: &login})
			}
			
			state := issue.State
			if state == "" {
				state = "open"
			}
			
			result = append(result, &github.Issue{
				Number:    &number,
				Title:     &title,
				Body:      &body,
				Labels:    githubLabels,
				Assignees: githubAssignees,
				State:     &state,
			})
		}
	}
	
	return result, nil
}

func (m *MockGitHubClient) AddLabels(ctx context.Context, number int, labels []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	if issue, ok := m.issues[number]; ok {
		// Add labels if they don't exist
		for _, newLabel := range labels {
			found := false
			for _, existingLabel := range issue.Labels {
				if existingLabel == newLabel {
					found = true
					break
				}
			}
			if !found {
				issue.Labels = append(issue.Labels, newLabel)
			}
		}
	}
	
	return nil
}

func (m *MockGitHubClient) RemoveLabel(ctx context.Context, number int, label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	if issue, ok := m.issues[number]; ok {
		// Remove label if it exists
		var newLabels []string
		for _, existingLabel := range issue.Labels {
			if existingLabel != label {
				newLabels = append(newLabels, existingLabel)
			}
		}
		issue.Labels = newLabels
	}
	
	return nil
}

func (m *MockGitHubClient) CreateComment(ctx context.Context, number int, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	comment := &MockComment{
		Body:      body,
		CreatedAt: time.Now(),
	}
	m.comments[number] = append(m.comments[number], comment)
	
	return nil
}

func (m *MockGitHubClient) AssignSelfIfNoAssignees(ctx context.Context, number int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	if issue, ok := m.issues[number]; ok {
		if len(issue.Assignees) == 0 {
			issue.Assignees = []string{"test-agent"}
		}
	}
	
	return nil
}

func (m *MockGitHubClient) AssignIssue(ctx context.Context, number int, assignees []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	if issue, ok := m.issues[number]; ok {
		issue.Assignees = assignees
	}
	
	return nil
}

func (m *MockGitHubClient) GetIssue(ctx context.Context, number int) (*github.Issue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	if issue, ok := m.issues[number]; ok {
		// Convert MockIssue to github.Issue
		num := issue.Number
		title := issue.Title
		body := issue.Body
		state := issue.State
		if state == "" {
			state = "open"
		}
		
		var githubLabels []*github.Label
		for _, labelName := range issue.Labels {
			name := labelName
			githubLabels = append(githubLabels, &github.Label{Name: &name})
		}
		
		var githubAssignees []*github.User
		for _, assignee := range issue.Assignees {
			login := assignee
			githubAssignees = append(githubAssignees, &github.User{Login: &login})
		}
		
		return &github.Issue{
			Number:    &num,
			Title:     &title,
			Body:      &body,
			Labels:    githubLabels,
			Assignees: githubAssignees,
			State:     &state,
		}, nil
	}
	
	return nil, fmt.Errorf("issue %d not found", number)
}

func (m *MockGitHubClient) CreateBranch(ctx context.Context, name string, fromRef string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	// Mock implementation - just track that it was called
	return nil
}

func (m *MockGitHubClient) ListPRs(ctx context.Context, state string) ([]*github.PullRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	var result []*github.PullRequest
	for _, pr := range m.prs {
		if state == "" || pr.State == state {
			number := pr.Number
			title := pr.Title
			body := pr.Body
			head := pr.Head
			base := pr.Base
			prState := pr.State
			
			result = append(result, &github.PullRequest{
				Number: &number,
				Title:  &title,
				Body:   &body,
				Head: &github.PullRequestBranch{
					Ref: &head,
				},
				Base: &github.PullRequestBranch{
					Ref: &base,
				},
				State: &prState,
			})
		}
	}
	
	return result, nil
}

func (m *MockGitHubClient) CreateIssue(ctx context.Context, title, body string, labels []string) (*github.Issue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	number := len(m.issues) + 1000 // Use high numbers to avoid conflicts
	issue := &MockIssue{
		Number: number,
		Title:  title,
		Body:   body,
		Labels: labels,
		State:  "open",
	}
	m.issues[number] = issue
	
	// Convert to github.Issue
	num := issue.Number
	issueTitle := issue.Title
	issueBody := issue.Body
	state := issue.State
	
	var githubLabels []*github.Label
	for _, labelName := range issue.Labels {
		name := labelName
		githubLabels = append(githubLabels, &github.Label{Name: &name})
	}
	
	return &github.Issue{
		Number: &num,
		Title:  &issueTitle,
		Body:   &issueBody,
		Labels: githubLabels,
		State:  &state,
	}, nil
}

func (m *MockGitHubClient) ListComments(ctx context.Context, number int) ([]*github.IssueComment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	comments := m.comments[number]
	var result []*github.IssueComment
	for _, comment := range comments {
		body := comment.Body
		createdAt := comment.CreatedAt
		id := int64(len(result) + 1)
		
		result = append(result, &github.IssueComment{
			ID:        &id,
			Body:      &body,
			CreatedAt: &github.Timestamp{Time: createdAt},
		})
	}
	
	return result, nil
}

func (m *MockGitHubClient) CreatePR(ctx context.Context, options ghub.PROptions) (*github.PullRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	prNumber := len(m.prs) + 1
	pr := &MockPR{
		Number: prNumber,
		Title:  options.Title,
		Body:   options.Body,
		Head:   options.Head,
		Base:   options.Base,
		State:  "open",
	}
	m.prs = append(m.prs, pr)
	
	// Convert to github.PullRequest
	number := pr.Number
	title := pr.Title
	body := pr.Body
	head := pr.Head
	base := pr.Base
	state := pr.State
	
	return &github.PullRequest{
		Number: &number,
		Title:  &title,
		Body:   &body,
		Head: &github.PullRequestBranch{
			Ref: &head,
		},
		Base: &github.PullRequestBranch{
			Ref: &base,
		},
		State: &state,
	}, nil
}

// SimpleClaudeClient is a minimal mock that focuses on testing agent behavior
type SimpleClaudeClient struct {
	mu               sync.RWMutex
	responses        map[string]string
	simulatedError   error
	callCount        int
	maxIterations    int
}

func NewSimpleClaudeClient() *SimpleClaudeClient {
	client := &SimpleClaudeClient{
		responses: make(map[string]string),
		maxIterations: 10,
	}
	
	// Set default responses
	client.responses["analyze"] = "## Analysis Complete\n\nThis issue requires implementing a new feature.\n\n## Implementation Plan\n\n1. Create new file\n2. Add tests\n3. Update docs"
	client.responses["too_complex"] = "COMPLEXITY_ASSESSMENT: TOO_COMPLEX\nThis issue requires decomposition."
	client.responses["implement"] = "Implementation completed successfully."
	
	return client
}

func (s *SimpleClaudeClient) SetResponse(input, response string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responses[input] = response
}

func (s *SimpleClaudeClient) SimulateError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simulatedError = err
}

func (s *SimpleClaudeClient) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simulatedError = nil
}

func (s *SimpleClaudeClient) SetMaxIterations(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxIterations = max
}

func (s *SimpleClaudeClient) GetCallCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.callCount
}

// For integration tests, we'll use a wrapper that satisfies the claude.Client interface
// but delegates to our simple mock for predictable behavior
type ClaudeClientWrapper struct {
	simple *SimpleClaudeClient
}

func NewClaudeClientWrapper(simple *SimpleClaudeClient) *ClaudeClientWrapper {
	return &ClaudeClientWrapper{simple: simple}
}

// MockStore implements state.Store interface for testing
type MockStore struct {
	mu             sync.RWMutex
	states         map[string]*state.AgentWorkState
	simulatedError error
}

func NewMockStore() *MockStore {
	return &MockStore{
		states: make(map[string]*state.AgentWorkState),
	}
}

func (m *MockStore) SimulateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulatedError = err
}

func (m *MockStore) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulatedError = nil
}

// Implement state.Store interface
func (m *MockStore) Save(ctx context.Context, agentState *state.AgentWorkState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	// Create a copy to avoid race conditions
	stateCopy := *agentState
	m.states[agentState.AgentType] = &stateCopy
	return nil
}

func (m *MockStore) Load(ctx context.Context, agentType string) (*state.AgentWorkState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	if state, ok := m.states[agentType]; ok {
		// Return a copy to avoid race conditions
		stateCopy := *state
		return &stateCopy, nil
	}
	
	return nil, nil
}

func (m *MockStore) Delete(ctx context.Context, agentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.simulatedError != nil {
		return m.simulatedError
	}

	delete(m.states, agentType)
	return nil
}

func (m *MockStore) List(ctx context.Context) ([]*state.AgentWorkState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.simulatedError != nil {
		return nil, m.simulatedError
	}

	var result []*state.AgentWorkState
	for _, state := range m.states {
		// Return copies to avoid race conditions
		stateCopy := *state
		result = append(result, &stateCopy)
	}
	
	return result, nil
}