package platform

// Mock implements Platform for testing.
// All operations are recorded and no system calls are made.
type Mock struct {
	TrashCalls      []string
	PickFolderPath  string
	PickFolderError error
	OpenBrowserURL  string
}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) SupportsTrash() bool { return true }

func (m *Mock) MoveToTrash(path string) error {
	m.TrashCalls = append(m.TrashCalls, path)
	return nil
}

func (m *Mock) ProtectedPrefixes() []string {
	return []string{
		"/System/",
		"/usr/",
		"/bin/",
	}
}

func (m *Mock) PickFolder() (string, error) {
	return m.PickFolderPath, m.PickFolderError
}

func (m *Mock) OpenBrowser(url string) error {
	m.OpenBrowserURL = url
	return nil
}
