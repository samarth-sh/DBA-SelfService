<script lang="ts">
    import { onMount } from 'svelte';
    let logs:Array<{ 
        requestID: number, 
        username: string, 
        serverIP: string, 
        requestType: string, 
        requestStatus: string, 
        message: string, 
        requestTime: string }> = [];
    let error: string | null = null;
    let username: string = '';
    let password: string = '';
    let showPassword: boolean = false;
    let loggedIn: boolean = false; 
    let user:string = '';

    function togglePasswordVisibility(event: Event): void {
        const checkbox = event.target as HTMLInputElement;
        showPassword = checkbox.checked;
    }

    async function handleSubmit() {
  try {
    const response = await fetch('http://localhost:8080/admin-login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username,
        password
      })
    });

    if (!response.ok) {
      throw new Error('Failed to login');
    }

    loggedIn = true;
    user = username;
    username = '';
    password = '';

    // Now fetch the logs from the redirected URL
    const logsResponse = await fetch('http://localhost:8080/getAllResetReq', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!logsResponse.ok) {
      throw new Error('Failed to fetch logs');
    }

    const data = await logsResponse.json();

    // Check if the data format is correct and update logs
    if (Array.isArray(data) && data.every(item =>
      typeof item.requestID === 'number' &&
      typeof item.username === 'string' &&
      typeof item.serverIP === 'string' &&
      typeof item.requestType === 'string' &&
      typeof item.requestStatus === 'string' &&
        typeof item.message === 'string' &&
      typeof item.requestTime === 'string')) {
      logs = data;
    } else {
      throw new Error('Invalid data format');
    }

    console.log("Fetched Logs:", logs);
  } catch (err: any) {
    error = err.message;
  }
}


    function handleLogout(): void {
        loggedIn = false;
    }
</script>

<main>
    {#if loggedIn}
    <div class="loginInfo">
    <h1>Admin Page</h1>
            <div class="user-info">
                    Logged in as: <p id="displayUsername">{user}</p>
            </div>
            <button id="LogoutBtn" on:click={handleLogout}>Logout</button>
    </div>
    {:else}
    <div class="form-content">
      <form on:submit|preventDefault={handleSubmit}>
        <label for="username">Username:</label>
        <input type="text" id="username" name="username" bind:value={username} placeholder="Enter admin username" required/>
    
        <label for="password">Password:</label>
        {#if showPassword}
        <input 
        type="text" 
        id="password" 
        name="password" 
        bind:value={password}
        placeholder="Enter admin password" required/>
        {:else}
        <input
        type="password" 
        id="password" 
        name="password" 
        bind:value={password}
        placeholder="Enter admin password" required/>
        {/if}
        <div class="checkbox">
            <label>
                <input
                    id="showPasswordBox"
                    type="checkbox"
                    bind:checked={showPassword}
                    on:change={togglePasswordVisibility}
                />
                Show Password
            </label>
        </div>
    
        <button id="loginBtn" type="submit">Login</button>
    </form>
</div>
{/if}
{#if error}
    <p style="color: red;">{error}</p>
{:else if loggedIn}
    <table>
        <thead>
            <tr>
                <th>Request ID</th>
                <th>Username</th>
                <th>Server IP</th>
                <th>Request Type</th>
                <th>Request Status</th>
                <th>Message</th>
                <th>Request Time</th>
            </tr>
        </thead>
        <tbody>
            {#each logs as log}
                <tr>
                    <td>{log.requestID}</td>
                    <td>{log.username}</td>
                    <td>{log.serverIP}</td>
                    <td>{log.requestType}</td>
                    <td>{log.requestStatus}</td>
                    <td>{log.message}</td>
                    <td>{log.requestTime}</td>
                </tr>
            {/each}
        </tbody>
    </table>
{/if}
</main>
<style>
    main{
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
        font-family: Poppins, sans-serif;
    }

    @media (max-width: 1000px) {
        table {
            font-size: 0.8rem;
        }
        th, td {
            padding: 6px;
        }
        .form-content form {
            max-width: 100%;
        }
        /* Add other elements here */
        main {
            padding: 20px;
        }
        h1 {
            font-size: 1.5rem;
        }
        .form-content {
            gap: 10px;
        }
        input {
            font-size: 0.8rem;
        }
        button {
            font-size: 0.8rem;
            padding: 0.4rem;
        }
    }
    @media (max-width: 400px) {
        table {
            font-size: 0.7rem;
        }
        th, td {
            padding: 4px;
        }
        h1 {
            font-size: 1.2rem;
        }
        .form-content form {
            max-width: 100%;
        }
        main {
            padding: 20px;
        }
        .form-content {
            gap: 10px;
        }
        input {
            font-size: 0.7rem;
        }
        button {
            font-size: 0.7rem;
            padding: 0.3rem;
        }
    }
   table {
        margin-top: 40px;
      width: 100%;
      border-collapse: collapse;
    } 
    th, td {
      border: 1px solid #b0afaf;
      padding: 8px;
    }
    th {
      background-color: #f5f3f3;
    }
    .checkbox {
        display: flex;
        flex-direction: row;
        align-items: center;
        margin-top: 10px;
        gap: 5px;
        font-size: 0.8rem;
    }
    #showPasswordBox{
        width: 14px;
        margin-top: 0.2rem auto;
    }
    .form-content{ 
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        max-width: 100%;
    }
    .form-content form{
        max-width: 450px;
        display: flex;
        flex-direction: column;
        gap: 14px;
        justify-content: center;
    }
    input{
        width: 100%;
        padding: 0.5rem;
        font-size: 0.9rem;
        border: 1px solid #ccc;
        border-radius: 4px;
    }
    input:focus{
        outline: none;
        border-color: #007bff;
    }
    button{
        width: 100%;
        padding: 0.5rem;
        font-size: 0.9rem;
        border: none;
        border-radius: 4px;
        background-color: #7e7f80;
        color: white;
        cursor: pointer;

    }
    #loginBtn{
        margin-top: 18px;
        text-align: center;
        padding: 10px 10px;
        border: none;
        border-radius: 4px;
    }

    /* .loginInfo {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        align-items: flex-start; 
        gap: 80px;
        margin-bottom: 20px;
    }

    .user-info {
        display: inline-flex;
        flex-direction: row;
        justify-content: flex-end; 
        align-items: center;
        margin-right: 10px;
    }

    #displayUsername {
        border: 1px solid #ccc;
        border-radius: 6px;
        padding: 4px 8px;
        margin-right: 10px;
    }

    #LogoutBtn {
        padding: 8px 12px;
        border: none;
        border-radius: 4px;
        background-color: #7e7f80;
        color: white;
        cursor: pointer;
    } */
    .loginInfo {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        padding: 10px;
        gap: 100px;
        background-color: #f8f9fa;
        border-bottom: 1px solid #ddd;
        width: 100%;
        box-sizing: border-box;
    }

    .loginInfo h1 {
        margin: 0;
        flex-shrink: 0;
    }

    .user-info {
        display: flex;
        flex-direction: row;
        align-items: center;
        /* justify-content: flex-end; */
        margin-left: auto;
    }

    .user-info p {
        margin: 0 10px;
    }

    #LogoutBtn {
        padding: 5px 10px;
        cursor: pointer;
        margin-left: 10px;
    }
</style>
  