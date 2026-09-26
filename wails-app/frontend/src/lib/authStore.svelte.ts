const initialState = typeof window !== 'undefined' && (window as any).__authStoreMock
  ? (window as any).__authStoreMock
  : { isAuthenticated: false, userEmail: "", licenseType: "" };

export const authStore = $state({
  isAuthenticated: initialState.isAuthenticated,
  userEmail: initialState.userEmail,
  licenseType: initialState.licenseType,

  login(email: string, token?: string) {
    const validDomains = ["@mckinsey.com", "@bcg.com", "@bain.com"];
    const isValid = validDomains.some(domain => email.toLowerCase().endsWith(domain));

    if (isValid) {
      this.isAuthenticated = true;
      this.userEmail = email;
      this.licenseType = "enterprise";
      return { success: true };
    } else {
      return { success: false, error: "No active enterprise license found for this domain." };
    }
  },

  logout() {
    this.isAuthenticated = false;
    this.userEmail = "";
    this.licenseType = "";
  }
});
