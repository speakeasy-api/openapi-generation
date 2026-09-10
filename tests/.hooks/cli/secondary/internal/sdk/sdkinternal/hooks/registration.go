package hooks

func initHooks(h *Hooks) {
	csh := &CustomSecurityHook{}
	h.registerBeforeRequestHook(csh)
}
