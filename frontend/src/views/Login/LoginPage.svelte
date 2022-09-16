<script context="module" lang="ts">
  export type Credentials = { username: string; password: string };
</script>

<script lang="ts">
  import { router } from "tinro";
  import Fa from "svelte-fa";
  import { faInfoCircle, faKey, faSpinner, faUserAlt } from "@icons";

  import { api, store } from "@lib";

  let loggingIn = false;
  let err = { cause: "", message: "" };

  const login = async (credentials: Credentials) => {
    try {
      loggingIn = true;
      const res = await api("/auth/login", "POST", credentials);
      store.auth.login(res.token);

      router.goto("/");
    } catch (error) {
      err.cause = error.cause;
      if (error.cause === "USER_NOT_FOUND") {
        err.message = "Username incorrect";
      } else if (error.cause === "PASSWORD_INCORRECT") {
        err.message = "Password incorrect";
      } else {
        err.message = "Something went wrong";
      }

      err = err;
    } finally {
      loggingIn = false;
    }
  };

  const handleSubmit = (e: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement }) => {
    const data = new FormData(e.currentTarget);
    const credentials = Object.fromEntries(data.entries()) as Credentials;
    login(credentials);
  };
</script>

<div class="container">
  <div class="header">
    <img src="https://source.boringavatars.com/bauhaus/70/MabyLink" alt="Logo" id="logo" />
    <h3>Log In to Rengoku</h3>
  </div>
  <div class="card">
    <form on:submit|preventDefault={handleSubmit}>
      <label for="username"> <Fa icon={faUserAlt} /> Username</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        type="text"
        name="username"
        id="username"
        required
        on:input={() => {
          err = { cause: "", message: "" };
        }}
        autofocus
        class:error={err.cause === "USER_NOT_FOUND"}
      />

      <label for="password"><Fa icon={faKey} /> Password</label>
      <input
        type="password"
        name="password"
        id="password"
        required
        on:input={() => {
          err = { cause: "", message: "" };
        }}
        class:error={err.cause === "PASSWORD_INCORRECT"}
      />

      <button type="submit">
        {#if loggingIn}
          <Fa icon={faSpinner} spin />
        {:else}
          Log In
        {/if}
      </button>
    </form>
  </div>
  <div class="message" class:error={err.message !== ""}>
    <Fa icon={faInfoCircle} size="sm" />
    {#if err.message != ""}
      {err.message}
    {:else}
      Input the credentials which you have set during initializing Rengoku
    {/if}
  </div>
</div>

<style>
  .container {
    position: absolute;
    margin-top: 6rem;
    right: 50%;
    transform: translate(50%);
    width: min-content;
  }
  .card {
    border-radius: 8px;
    border: 0.5px solid var(--background);
    padding: 1em;
    background-color: var(--background-alt);

    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .header {
    width: 100%;
    text-align: center;
    padding-bottom: 2em;
  }
  label {
    margin-bottom: 0.7em;
  }
  input {
    border: 0.5px solid var(--background-body);
    margin: 0;
    margin-bottom: 1.2em;
  }
  button {
    margin-top: 1em;
    width: 100%;
  }
  .message {
    border: 0.5px dashed var(--text-muted);
    border-radius: 8px;
    padding: 1em;
    margin-top: 1em;
    color: var(--text-muted);
  }
  .error {
    color: rgb(231, 126, 126);
    border-color: rgb(231, 126, 126);
  }
</style>
