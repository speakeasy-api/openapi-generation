package hooks

func initHooks(h *Hooks) {
	th := &TestHook{}
	csh := &CustomSecurityHook{}

	h.registerSDKInitHook(th)

	h.registerBeforeRequestHook(th)
	h.registerBeforeRequestHook(csh)

	h.registerAfterSuccessHook(th)

	h.registerAfterErrorHook(th)
}
