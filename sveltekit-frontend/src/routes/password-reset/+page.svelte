<script lang="ts">
	import { tick } from 'svelte';
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

    function preventCopyPaste(event: Event){
        event.preventDefault();
    }

    let passMeetsRequirements: boolean = false;
    let passwordsMatch: boolean = false;

    $: passwordsMatch = newPassword === confirmPassword && newPassword !== '';

    function validatePassword(password: string): boolean {
        return password.length >= 8 && 
        password.length <= 16 &&
        !password.includes(username) &&
        !password.includes(oldPassword) &&
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$#!%*?&])[A-Za-z\d@$#!%*?&]{8,16}$/.test(password);
    }

    $: passMeetsRequirements = validatePassword(newPassword);

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
        console.log('Sending request:', { username, oldPassword, newPassword, serverIP });
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
            <h2>Password Reset Form</h2>
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
        <div class="page-content">
         
            <form on:submit|preventDefault={updatePassword}>


            <label for="username">Username</label>
            <input id="username" bind:value={username} placeholder="Enter your username" required>

            <label for="serverIP">Server IP</label>
            <input id="serverIP" bind:value={serverIP} placeholder="Enter Server IP (10.xxx.xxx.xxx)" required>

            <label for="oldPassword">Current Password</label>
            {#if showPassword}
                <input
                type="text"
                id="oldPassword"
                bind:value={oldPassword}
                placeholder="Enter your current password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {:else}
                <input
                type="password"
                id="oldPassword"
                bind:value={oldPassword}
                placeholder="Enter your current password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {/if}

            <label for="newPassword">New Password</label>
            <div class="password-input">
            {#if showPassword}
                <input
                type="text"
                id="newPassword"
                bind:value={newPassword}
                placeholder="Enter your new password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {:else}
                <input
                type="password"
                id="newPassword"
                bind:value={newPassword}
                placeholder="Enter your new password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {/if}
            {#if passMeetsRequirements}
            <span class="tick-mark" class:visible={passMeetsRequirements}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="5" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
            </span>            
            {/if}
            </div>

            <label for="confirmPassword">Confirm Password</label>
            <div class="password-input">
            {#if showPassword}
                <input
                type="text"
                id="confirmPassword"
                bind:value={confirmPassword}
                placeholder="Re-type the new password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {:else}
                <input
                type="password"
                id="confirmPassword"
                bind:value={confirmPassword}
                placeholder="Re-type the new password"
                on:copy={preventCopyPaste}
                on:cut={preventCopyPaste}
                on:paste={preventCopyPaste}
                />
            {/if}
            {#if passwordsMatch}
            <span class="tick-mark" class:visible={passwordsMatch}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="5" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
            </span>            
            {/if}
            </div>
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
</div>

</main>
    
<style>
    main{
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
        font-family: Poppins, sans-serif;
    }
    @media (max-width: 800px) {
        main {
            padding: 10px;
            max-width: 100%;
        }
        h1{
            font-size: 1.5rem;
        }
        h2{
            font-size: 1.2rem;
        }
        h3{
            font-size: 1rem;
        }
        input, button {
            font-size: 0.8rem;
            padding: 5px;
        }
        
    }
    @media (min-width:225px){
        main {
            padding: 12px;
            max-width: 100%;
            margin-left: 35px;
            margin-right: 35px;
            margin-top: 0 auto;
        }
        h1 {
            font-size: 1.5rem;
        }
        h2, h3{
            font-size: 1rem;
        }
        .inputform{
            font-size: smaller;
            display: flex;
            flex-direction: column;
            gap: 2px;
            
        }
        input{
            font-size: 0.9rem;
            padding: 5px;
        }
        .password-requirements{
            margin: 5px;
            padding: 0.1rem auto;
        }
        ul, li {
            font-size: 0.65rem;
        }
        form{
            width: 100%;
            margin: 0.2rem;
        }
        .page-content{
            display: flex;
            flex-direction: row;
            gap: 5px;
            align-items: center;
            justify-content: center;
        }


    }
    .inputform{
        margin-left: 20px;
        margin-right: 20px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        max-width: 800px;
        justify-content: center;
        align-items: center;
    }
    form{
        width: 600px;
        margin: 0.3rem auto;
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
        margin-top: 4px;
        margin-bottom: 8px;
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
        width: 14px;
        margin-top: 0.2rem auto;
    }

    .page-content{
        display: flex;
        flex-direction: row;
        gap: 50px;
        justify-content: center;
        align-items: center;
        max-width: 100%;

    }
    .page-content form{
        flex: 0 0 40%;
        margin-right: 40px;
        
    }
    .page-content .password-requirements{
        flex: 0 0 40%;
        padding: 20px 10px;
        border: 1px solid #787676;
        border-radius: 8px;
        background-color: #efebeb;
        font-size: 0.9rem;
        font-style: italic;
        margin: 0.2rem auto;
        max-width: 100%;
    }
    .page-content .password-requirements h4{
        margin-top: 0;
    }
    .password-input{
        position: relative;
        width: 100%;
    }
    .password-input input{
        width: 100%;
        padding: 0.5rem;
    }
    .tick-mark{
        position: absolute;
        top: 50%;
        right: 5px;
        transform: translateY(-50%);
        width: 20px;
        height: 20px;
        color: #4CAF50;
        font-size: 1.2rem;
        transition: opacity 0.8s ease;
        opacity: 0;
    }

    .tick-mark.visible{
        opacity: 1;
    }
    .tick-mark svg{
        width: 100%;
        height: 100%;
    }

    </style>

