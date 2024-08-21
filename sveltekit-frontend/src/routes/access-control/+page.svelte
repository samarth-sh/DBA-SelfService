 <script lang="ts">
    let username: string = '';
    let database: string = '';
    let accessLevel: string = '';
    let reason: string = '';
    let errorMessage: string = '';
    let successMessage: string = '';
    
    async function handleAccessReq() {

        errorMessage = '';
        successMessage = '';
        
        const response = await fetch(`http://localhost:8080/access-request`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username,
                database,
                accessLevel,
                reason
            })
        });
        console.log('Sending request:', { username, database, accessLevel, reason });
        const result = await response.json();
        if (!response.ok) {
            if(result.error){
                errorMessage = result.error;
            }
            else{
                errorMessage = 'Failed to submit request';
            }
            return;
        }
        else{
            username = '';
            database = '';
            accessLevel = '';
            reason = '';
            successMessage = result.message || 'Request submitted successfully';
        }
    }
 </script>
 <main>
    <div class="headers">
        <h1>DBA Self Service</h1>
        <h2>Access Request Form</h2>
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
<div class="form-container">
<form on:submit|preventDefault={handleAccessReq}>

    <div class="input-container">
        <label for="username">Username: </label>
        <input id="username" bind:value={username} required />
    </div>

    <div class="input-container">
        <label for="database">Database: </label>
        <input id="database" bind:value={database} required />
    </div>

    <div class="input-container">
        <label for="accessLevel">Access Level: </label>
        <select bind:value={accessLevel} required>
            <option value="" disabled selected hidden>Select Access Level</option>
            <option value="read">Read</option>
            <option value="write">Write</option>
            <option value="admin">Admin</option>
        </select>
    </div>

    <div class="input-container">
        <label for="reason">Reason for Access: </label>
        <textarea id="reason" bind:value={reason}></textarea>
    </div>

    <button type="submit">Submit Request</button>
</form>
</div>

</main>
<style>
main{
    max-width: 800px;
    margin: 10px auto;
    padding: 20px;
    font-family: Poppins, sans-serif;
}
.headers{
        text-align: center;
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
    @media (max-width:550px){
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
        input{
            font-size: 0.9rem;
            padding: 5px;
        }
        form{
            width: 100%;
            margin: 0.2rem;
        }
    }
.form-container{
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 10px;
    height: 60vh;
}
form{
    display: flex;
    flex-direction: column;
    gap: 10px;
    justify-content: center;
    align-items: center;
    margin: 0.2rem;
}
input{
    font-size: 0.9rem;
    padding: 5px;
    border: 1px solid #797979;
    border-radius: 5px;
}
input:focus{
        outline: none;
        border-color: #006aff;
        
    }
.input-container {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    margin-bottom: 20px;
}
.input-container input,
.input-container select,
.input-container textarea {
    width: 100%;
    padding: 9px;
    font-size: 0.9rem;
    /* border: 1px solid ; */
    border-radius: 4px;
    outline: none;
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
        background-color: #004896;
    }
}
    .messages{
        margin-top: 14px;
        margin-bottom: 8px;
        font-size: 0.9rem;
        font-weight: bold;
        text-transform: uppercase;
    }
    #message{
        margin-top: 14px;
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

</style>