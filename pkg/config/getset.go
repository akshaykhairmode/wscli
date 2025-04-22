package config

// func (c *Flag) IsPerf() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isPerf
// }

// func (c *Flag) SetPerf(isPerf bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isPerf = isPerf
// }

// func (c *Flag) GetPerfConfig() Perf {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.perf
// }

// func (c *Flag) SetPerfConfig(perf Perf) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.perf = perf
// }

// func (c *Flag) GetConnectURL() string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.connectURL
// }

// func (c *Flag) SetConnectURL(connectURL string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.connectURL = connectURL
// }

// func (c *Flag) GetAuth() string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.auth
// }

// func (c *Flag) SetAuth(auth string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.auth = auth
// }

// func (c *Flag) GetHeaders() []string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.headers
// }

// func (c *Flag) SetHeaders(headers []string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.headers = headers
// }

// func (c *Flag) GetOrigin() string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.origin
// }

// func (c *Flag) SetOrigin(origin string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.origin = origin
// }

// func (c *Flag) GetExecute() []string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.execute
// }

// func (c *Flag) SetExecute(execute []string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.execute = execute
// }

// func (c *Flag) GetWait() time.Duration {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.wait
// }

// func (c *Flag) SetWait(wait time.Duration) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.wait = wait
// }

// func (c *Flag) GetSubProtocol() []string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.subProtocol
// }

// func (c *Flag) SetSubProtocol(subProtocol []string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.subProtocol = subProtocol
// }

// func (c *Flag) GetProxy() string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.proxy
// }

// func (c *Flag) SetProxy(proxy string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.proxy = proxy
// }

// func (c *Flag) ShowPingPong() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.showPingPong
// }

// func (c *Flag) SetShowPingPong(showPingPong bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.showPingPong = showPingPong
// }

// func (c *Flag) IsSlash() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isSlash
// }

// func (c *Flag) SetIsSlash(isSlash bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isSlash = isSlash
// }

// func (c *Flag) SkipCertificateCheck() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.noCertificateCheck
// }

// func (c *Flag) SetNoCertificateCheck(noCertificateCheck bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.noCertificateCheck = noCertificateCheck
// }

// func (c *Flag) IsShowVersion() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.version
// }

// func (c *Flag) SetVersion(version bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.version = version
// }

// func (c *Flag) IsVerbose() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.verbose
// }

// func (c *Flag) SetVerbose(verbose bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.verbose = verbose
// }

// func (c *Flag) IsNoColor() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.noColor
// }

// func (c *Flag) SetNoColor(noColor bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.noColor = noColor
// }

// func (c *Flag) ShowResponseHeaders() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.shouldShowResponseHeaders
// }

// func (c *Flag) SetResponse(response bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.shouldShowResponseHeaders = response
// }

// func (c *Flag) IsJSONPrettyPrint() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isJSONPrettyPrint
// }

// func (c *Flag) SetJSONPrettyPrint(jSONPrettyPrint bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isJSONPrettyPrint = jSONPrettyPrint
// }

// func (c *Flag) IsBinary() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isBinary
// }

// func (c *Flag) SetIsBinary(isBinary bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isBinary = isBinary
// }

// func (c *Flag) IsHelp() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.help
// }

// func (c *Flag) SetHelp(help bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.help = help
// }

// func (c *Flag) IsStdin() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isSTDin
// }

// func (c *Flag) SetStdin(stdin bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isSTDin = stdin
// }

// func (c *Flag) IsGzipResponse() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isGzipResponse
// }

// func (c *Flag) IsStdOut() bool {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.isStdOut
// }

// func (c *Flag) SetGzipResponse(gzipr bool) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.isGzipResponse = gzipr
// }

// func (c *Flag) GetTLS() TLS {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.tls
// }

// func (c *Flag) SetTLS(tls TLS) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.tls = tls
// }

// // Getters and Setters for TLS

// func (t *TLS) GetCA() string {
// 	return t.CA
// }

// func (t *TLS) SetCA(ca string) {
// 	t.CA = ca
// }

// func (t *TLS) GetCert() string {
// 	return t.Cert
// }

// func (t *TLS) SetCert(cert string) {
// 	t.Cert = cert
// }

// func (c *Flag) GetPingInterval() time.Duration {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.pingInterval
// }

func (c *Flag) ShouldProcessAsCmd() bool {
	if len(Flags.Execute) > 0 && Flags.Wait > 0 {
		return true
	}

	if Flags.IsSTDin {
		return true
	}

	return false
}

// func (c *Flag) GetPrintInterval() time.Duration {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.printOutputInterval
// }

// func (c *Flag) SetPrintInterval(dur time.Duration) {
// 	c.printOutputInterval = dur
// }

// func (c *Flag) GetPerfOutfile() string {
// 	c.mux.RLock()
// 	defer c.mux.RUnlock()
// 	return c.perf.LogOutFile
// }

// func (c *Flag) SetPerfOutfile(file string) {
// 	c.mux.Lock()
// 	defer c.mux.Unlock()
// 	c.perf.LogOutFile = file
// }
