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
      <h1 class="login-title">Realm of Eternity</h1>
      <p class="login-subtitle">Enter the world of magic and adventure</p>
    </div>

    <div class="login-buttons">
      <button
        class="login-btn google-btn"
        on:click={handleGoogleLogin}
        disabled={guestLoading || googleLoading}
      >
        {#if googleLoading}
          <span class="btn-spinner">⏳</span>
        {:else}
          <span class="btn-icon">🔑</span>
        {/if}
        Sign in with Google
      </button>

      <div class="divider">OR</div>

      <button
        class="login-btn guest-btn"
        on:click={handleGuestLogin}
        disabled={guestLoading || googleLoading}
      >
        {#if guestLoading}
          <span class="btn-spinner">⏳</span>
        {:else}
          <span class="btn-icon">👤</span>
        {/if}
        Play as Guest
      </button>
    </div>

    {#if errorMessage}
      <div class="error-message">
        {errorMessage}
      </div>
    {/if}

    <p class="login-note">
      ✨ Your progress will be saved locally on this device
    </p>
  </div>
</div>

<style>
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: linear-gradient(135deg, var(--color-bg-dark) 0%, var(--color-danger-dim) 100%);
    padding: 20px;
    font-family: var(--font-body);
  }

  .login-card {
    background: var(--color-bg-panel);
    border: 2px solid var(--color-gold-dim);
    border-radius: 16px;
    padding: 48px 32px;
    max-width: 400px;
    width: 100%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .login-header {
    text-align: center;
    margin-bottom: 32px;
  }

  .login-title {
    font-family: var(--font-heading);
    font-size: 32px;
    color: var(--color-gold-bright);
    margin: 0 0 8px;
    text-shadow: 0 2px 8px rgba(212, 175, 55, 0.3);
  }

  .login-subtitle {
    font-size: 14px;
    color: var(--color-text-muted);
    margin: 0;
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
    border-radius: 8px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .login-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .google-btn {
    background: #4285f4;
    color: white;
  }

  .google-btn:hover:not(:disabled) {
    background: #3367d6;
    transform: translateY(-2px);
  }

  .guest-btn {
    background: var(--color-gold-dim);
    color: #000;
  }

  .guest-btn:hover:not(:disabled) {
    background: var(--color-gold-bright);
    transform: translateY(-2px);
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
