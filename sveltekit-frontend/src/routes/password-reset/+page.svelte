<script lang="ts">
    let username: string = '';
    let oldPassword: string = '';
    let newPassword: string = '';
    let confirmPassword: string = '';
    let serverIP: string = '';
    let errorMessage: string = '';
    let successMessage: string = '';

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
        if(!newPassword.match(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,16}$/)) {
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
    <div class="headers">
        <h1>Password Reset</h1>

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
    <div class="inputform">        
        <form on:submit|preventDefault={updatePassword}>
            <label for="username">Username</label>
            <input id="username" bind:value={username} placeholder="Enter your username" required>

            <label for="serverIP">Server IP</label>
            <input id="serverIP" bind:value={serverIP} placeholder="Enter Server IP (10.xxx.xxx.xxx)" required>

            <label for="oldPassword">Old Password</label>
            <input id="oldPassword" bind:value={oldPassword} placeholder="Enter your current password" type="password" required>

            <label for="newPassword">New Password</label>
            <input id="newPassword" bind:value={newPassword} placeholder="Enter your new password" type="password" required>

            <label for="confirmPassword">Confirm Password</label>
            <input id="confirmPassword" bind:value={confirmPassword} placeholder="Re-type the new password" type="password" required>

            <button id="upd" type="submit">Update Password</button>
        </form>
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
        form{
            display: flex;
            flex-direction: column;
            gap: 10px;
        }
        input{
            padding: 10px;
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
            text-align: center;
            margin-top: 10px;
            margin-bottom: 10px;
            font-size: 0.9rem;
            font-weight: bold;
            color: #007bff;
            text-transform: uppercase;
        }
        button{
            padding: 10px;
            border: none;
            border-radius: 4px;
            background-color: #007bff;
            color: white;
            cursor: pointer;
        }
        #upd{
            margin-top: 18px;
            text-align: center;
            padding: 10px 10px;
            border: none;
            border-radius: 4px;
        }
        button:hover{
            background-color: #0056b3;
        }
        #message{
            margin-top: 10px;
            font-size: 0.9rem;
            font-weight: bold;
            text-transform: uppercase;
            text-align: center;
            color: #007bff;
        }
        .error{
            color: red;
        }
        .success{
            color: green;
        }

    </style>
