<script lang="ts">
    let username: string = '';
    let oldPassword: string = '';
    let newPassword: string = '';
    let confirmPassword: string = '';
    let serverIP: string = '';
    let errorMessage: string = '';
    let successMessage: string = '';
    let showPassword: boolean = false;

    function togglePasswordVisibility(event: Event): void {
        const checkbox = event.target as HTMLInputElement;
        showPassword = checkbox.checked;
    }

    function validateInput(){
        if(newPassword !== confirmPassword) {
            errorMessage = 'Passwords do not match';
            return;
        }
        if (newPassword.length < 8 || newPassword.length > 16) {
            errorMessage = 'Password must be between 8 and 16 characters long';
            return;
        }
        if(!serverIP.match(/^10\.\d{1,3}\.\d{1,3}\.\d{1,3}$/)) {
            errorMessage = 'Invalid Server IP';
            return;
        }
        if (newPassword.includes(username)) {
            errorMessage = 'Password cannot contain username';
            return;
        }
        if (newPassword.includes(oldPassword)) {
            errorMessage = 'Password cannot contain old password';
            return;
        }
        if(!newPassword.match(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$#!%*?&])[A-Za-z\d@$#!%*?&]{8,16}$/)) {
            errorMessage = 'Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character';
            return;
        }
        return true;
    }

    async function updatePassword(){
        if(!validateInput()){
            return;
        }

        errorMessage = '';
        successMessage = '';

        const response = await fetch(`http://localhost:8080/update-password`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username,
                oldPassword,
                newPassword,
                serverIP
            })
        });
        const result = await response.json();
        if (!response.ok) {
            if(result.error){
                errorMessage = result.error;
            }
            else{
                errorMessage = 'Failed to update password';
            }
            return;
        }
        else{
            username = '';
            oldPassword = '';
            newPassword = '';
            confirmPassword = '';
            serverIP = '';
            successMessage = result.message || 'Password updated successfully';
        }
    }
    </script>
    
<main>

    <div class="inputform">  
        <div class="headers">
            <h1>DBA Self Service</h1>
            <h2>Password Reset</h2>
            <h3>Enter your credentials</h3>
        </div>
        <div class="messages">
            {#if errorMessage}
                <p id="message" class="error">{errorMessage}</p>
            {/if}
            {#if successMessage}
                <p id="message" class="success">{successMessage}</p>
            {/if}
        </div>
        <form on:submit|preventDefault={updatePassword}>
        
        <label for="username">Username</label>
        <input id="username" bind:value={username} placeholder="Enter your username" required>
           
        <label for="serverIP">Server IP</label>
        <input id="serverIP" bind:value={serverIP} placeholder="Enter Server IP (10.xxx.xxx.xxx)" required>
            
        <label for="oldPassword">Old Password</label>
            {#if showPassword}
            <input
            type="text"
            id="oldPassword"
            bind:value={oldPassword}
            placeholder="Enter your old password"
            />
            {:else}
            <input
            type="password"
            id="oldPassword"
            bind:value={oldPassword}
            placeholder="Enter your old password"
            />
            {/if}

            <label for="newPassword">New Password</label>
            {#if showPassword}
            <input
            type="text"
            id="newPassword"
            bind:value={newPassword}
            placeholder="Enter your new password"
            />
            {:else}
            <input
            type="password"
            id="newPassword"
            bind:value={newPassword}
            placeholder="Enter your new password"
            />
            {/if}

            <label for="confirmPassword">Confirm Password</label>
            {#if showPassword}
            <input
            type="text"
            id="confirmPassword"
            bind:value={confirmPassword}
            placeholder="Re-type the new password"
            />
            {:else}
            <input
            type="password"
            id="confirmPassword"
            bind:value={confirmPassword}
            placeholder="Re-type the new password"
            />
            {/if}
           
            <div class="checkbox">
                <label>
                <input
                id="showPasswordBox"
                type="checkbox"
                bind:checked={showPassword}
                on:change={togglePasswordVisibility}
                />
                Show Passwords
                </label>
            </div>
            <button id="upd" type="submit">Update Password</button>
        </form>
    </div>
    <div class="password-requirements">
        <h4>Password Requirements</h4>
        <ul>
        <li>Password must contain at least one
         <ul>
            <li>Uppercase letter</li>
            <li>Lowercase letter</li>
            <li>Number</li>
            <li>Special character</li>
            </ul>
         </li>
         <li>Password must be between 8 and 16 characters long</li>
         <li>Password cannot contain username or old password</li>
         <li>Password can contain the following special characters: @$#!%*?&</li>
         </ul>
    </div>
</main>
    
<style>
    main{
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
        font-family: Poppins, sans-serif;
    }
    @media (max-width: 600px) {
        main {
            padding: 10px;
            max-width: 100%;
        }
        h1{
            font-size: 1.5rem;
        }
        h3{
            font-size: 1rem;
        }
        input, button {
            font-size: 0.8rem;
            padding: 5px;
        }
    }
    .inputform{
        display: flex;
        flex-direction: column;
        gap: 4px;
    }
    form{
        width: 600px;
        margin: 0.1rem auto;
        display: flex;
        flex-direction: column;
        gap: 10px;
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
    .headers{
        text-align: center;
    }
    .messages{
        margin-top: 10px;
        margin-bottom: 10px;
        font-size: 0.9rem;
        font-weight: bold;
        text-transform: uppercase;
    }
  input[type="text"],
    input[type="password"] {
        flex: 1;
    }
    button{
        width: 100%;
        padding: 0.5rem;
        font-size: 0.9rem;
        border: none;
        border-radius: 4px;
        background-color: #007bff;
        color: white;
        cursor: pointer;

        &:hover{
            background-color: #0056b3;
        }
    }
    #upd{
        margin-top: 18px;
        text-align: center;
        padding: 10px 10px;
        border: none;
        border-radius: 4px;
    }
    #message{
        margin-top: 10px;
        font-size: 0.9rem;
        font-weight: bold;
        text-transform: uppercase;
        text-align: center;
    }
    .error{
        color: red;
    }
    .success{
        color: green;
    }
    .checkbox {
        display: flex;
        flex-direction: row;
        align-items: center;
        margin-top: 10px;
        gap: 5px;
        font-size: 0.9rem;
    }
    #showPasswordBox{
        width: 10px;
        margin-top: 10px;
    }
    .password-requirements{
        margin-top: 20px;
        padding: 10px 20px;
        border: 1px solid #ccc;
        border-radius: 8px;
        background-color: #efebeb;
        font-size: 0.9rem;
        opacity: 0.8;
        font-style: italic;
        max-width: 100%;
        max-height: 300vh;
    }
    </style>

