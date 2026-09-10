package extensions

type SSESentinel struct {
	DataValue string
}

func (e *Extensions) HandleSSESentinelExtension(extensions OAExtensions) (*SSESentinel, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	dataVal, err := getExtensionValue(e.GetResolvedName(ExtSSESentinel), extensions, "")
	if err != nil {
		return nil, err
	}

	return &SSESentinel{
		DataValue: dataVal,
	}, nil
}

func (e *Extensions) HandleSSESentinelExtensionString(extensions OAExtensions) string {
	sentinel, err := e.HandleSSESentinelExtension(extensions)
	if err != nil || sentinel == nil {
		return ""
	}
	return sentinel.DataValue
}
