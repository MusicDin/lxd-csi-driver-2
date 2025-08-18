package e2e

// PrettyName returns a string consisting of resource's namespace, and name.
// If the namespace is empty, it returns only the name.
// If the generateName is provided, it appends "xxxxx" to it to indicate
// that the name is generated.
func PrettyName(namespace string, generateName string, name string) string {
	if namespace != "" {
		namespace += "/"
	}

	if name != "" {
		return namespace + name
	}

	// xxxxx is a placeholder for the generated suffix.
	return namespace + generateName + "xxxxx"
}
