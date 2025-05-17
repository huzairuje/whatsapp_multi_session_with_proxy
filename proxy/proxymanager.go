package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Manager orchestrates proxy assignments for users.
// It handles balanced distribution and dynamic balancing.
type Manager struct {
	mu      sync.Mutex
	proxies []Proxy        // slice of all proxies
	assign  map[string]int // userJID -> proxy index
	loads   []int          // proxy index -> number of users assigned
}

// LoadFromFile reads a proxy-list file (each line "baseURL:port:username:password"),
// loads them into the Manager, and balances assignments.
// Parsing is done off-lock, then ResetProxies runs under lock to avoid races.
func (m *Manager) LoadFromFile(path string) error {
	// parse file without holding lock
	list, err := loadProxyList(path)
	if err != nil {
		return err
	}
	// atomically reset under lock
	m.ResetProxies(list)
	return nil
}

// loadProxyList parses the given file into a slice of Proxy.
func loadProxyList(path string) ([]Proxy, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening proxy list: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var list []Proxy
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid proxy format on line %d: %q", lineNum, line)
		}
		list = append(list, Proxy{
			BaseURL:  parts[0],
			Port:     parts[1],
			Username: parts[2],
			Password: parts[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning proxy list: %w", err)
	}
	return list, nil
}

// ResetProxies replaces the proxy list and reassigns all current users
// to balance loads evenly across the new list.
func (m *Manager) ResetProxies(list []Proxy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.proxies = make([]Proxy, len(list))
	copy(m.proxies, list)

	// reset load counts
	m.loads = make([]int, len(list))
	// reassign existing users to the least-loaded proxies
	for user := range m.assign {
		idx := m.pickMinLoaded()
		m.assign[user] = idx
		m.loads[idx]++
	}
}

// AddUser assigns a proxy to the given userJID if not already assigned.
// Returns the assigned Proxy or an error if no proxies are available.
func (m *Manager) AddUser(userJID string) (Proxy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// return existing assignment
	if idx, ok := m.assign[userJID]; ok {
		return m.proxies[idx], nil
	}

	if len(m.proxies) == 0 {
		return Proxy{}, errors.New("no proxies available")
	}

	idx := m.pickMinLoaded()
	m.assign[userJID] = idx
	m.loads[idx]++
	return m.proxies[idx], nil
}

// GetUser retrieves the assigned proxy for userJID.
// The bool indicates if an assignment exists.
func (m *Manager) GetUser(userJID string) (Proxy, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx, ok := m.assign[userJID]
	if !ok {
		return Proxy{}, false
	}
	return m.proxies[idx], true
}

// RemoveUser unassigns the proxy for userJID and decrements its load.
func (m *Manager) RemoveUser(userJID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if idx, ok := m.assign[userJID]; ok {
		m.loads[idx]--
		delete(m.assign, userJID)
	}
}

// pickMinLoaded returns the index of the proxy with the smallest load.
// In case of ties, the lowest index is chosen.
func (m *Manager) pickMinLoaded() int {
	// assume at least one proxy exists
	minIdx := 0
	minLoad := m.loads[0]
	for i, load := range m.loads {
		if load < minLoad {
			minLoad = load
			minIdx = i
		}
	}
	return minIdx
}
