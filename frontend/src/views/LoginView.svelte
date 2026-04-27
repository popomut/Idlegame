<script>
  import { navigateTo } from '../stores/navigation.js';
  import { isAuthenticated } from '../stores/auth.js';
  import { API_BASE_URL } from '../services/api.js';

  let guestLoading = false;
  let googleLoading = false;
  let errorMessage = '';

  async function handleGuestLogin() {
    guestLoading = true;
    errorMessage = '';
    
    try {
      // Call backend to create guest account
      const response = await fetch(`${API_BASE_URL}/api/auth/guest`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include', // Important: send cookies
        body: JSON.stringify({})
      });
      
      if (!response.ok) {
        throw new Error('Failed to create guest account');
      }

      console.log('Guest login successful. Cookies should be set.');
      
      // Token is now in httpOnly cookie - no need to store it
      // Just update auth state
      isAuthenticated.set(true);
      
      // Redirect to home
      navigateTo('home');
    } catch (error) {
      errorMessage = `❌ ${error.message}`;
      console.error(error);
    } finally {
      guestLoading = false;
    }
  }

  function handleGoogleLogin() {
    googleLoading = true;
    // TODO: Implement Google OAuth
    errorMessage = '🔄 Google login coming soon...';
    setTimeout(() => {
      googleLoading = false;
    }, 2000);
  }
</script>

<div class="login-container">
  <div class="login-card">
    <div class="login-header">
      <h1 class="login-title">TOXIC PROTOCOL</h1>
      <p class="login-subtitle">&#x2623;&#xFE0F; Survive the contaminated zone</p>
    </div>

    <div class="login-buttons">
      <button
        class="login-btn google-btn"
        on:click={handleGoogleLogin}
        disabled={guestLoading || googleLoading}
      >
        {#if googleLoading}
          <span class="btn-spinner">&#x23F3;</span>
        {:else}
          <span class="btn-icon">&#x1F511;</span>
        {/if}
        Intel Access via Google
      </button>

      <div class="divider">— OR —</div>

      <button
        class="login-btn guest-btn"
        on:click={handleGuestLogin}
        disabled={guestLoading || googleLoading}
      >
        {#if guestLoading}
          <span class="btn-spinner">&#x23F3;</span>
        {:else}
          <span class="btn-icon">&#x1FA96;</span>
        {/if}
        Deploy as Recruit
      </button>
    </div>

    {#if errorMessage}
      <div class="error-message">
        {errorMessage}
      </div>
    {/if}

    <p class="login-note">
      &#x2622;&#xFE0F; Mission data stored in secure field cache
    </p>
  </div>
</div>

<style>
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background:
      linear-gradient(135deg, var(--color-bg) 0%, #0a1a0a 60%, #0d1f0a 100%);
    padding: 20px;
    font-family: var(--font-body);
  }

  .login-card {
    background: var(--color-bg-panel);
    border: 2px solid var(--color-magic);
    border-radius: 4px;
    padding: 48px 32px;
    max-width: 400px;
    width: 100%;
    box-shadow: 0 0 40px rgba(42, 158, 42, 0.15), 0 20px 60px rgba(0, 0, 0, 0.6);
  }

  .login-header {
    text-align: center;
    margin-bottom: 32px;
  }

  .login-title {
    font-family: var(--font-heading);
    font-size: 36px;
    color: var(--color-magic-bright);
    margin: 0 0 8px;
    letter-spacing: 4px;
    text-shadow: 0 0 20px rgba(61, 200, 61, 0.5);
    text-transform: uppercase;
  }

  .login-subtitle {
    font-size: 13px;
    color: var(--color-text-muted);
    margin: 0;
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  .login-buttons {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 24px;
  }

  .login-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 14px 20px;
    border: none;
    border-radius: 3px;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  .login-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .google-btn {
    background: #2a5a9e;
    color: white;
  }

  .google-btn:hover:not(:disabled) {
    background: #1e4a8e;
    transform: translateY(-1px);
  }

  .guest-btn {
    background: var(--color-magic);
    color: #000;
  }

  .guest-btn:hover:not(:disabled) {
    background: var(--color-magic-bright);
    transform: translateY(-1px);
  }

  .btn-icon {
    font-size: 20px;
  }

  .btn-spinner {
    animation: spin 1s linear infinite;
  }

  .divider {
    text-align: center;
    color: var(--color-text-muted);
    font-size: 12px;
    margin: 8px 0;
  }

  .error-message {
    background: rgba(220, 38, 38, 0.1);
    border-left: 3px solid var(--color-danger-bright);
    padding: 12px;
    border-radius: 4px;
    color: var(--color-danger-bright);
    font-size: 14px;
    margin-bottom: 16px;
  }

  .login-note {
    text-align: center;
    font-size: 12px;
    color: var(--color-text-muted);
    margin: 0;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
