<script>
    import { goto } from '$app/navigation';

    function goToPasswordReset() {
        goto('/password-reset');
    }

    function goToAccessControl() {
        goto('/access-control');
    }
</script>

<div class="mainContainer">
<h1>DBA Self Service Portal</h1>

<div class="options">
    <button on:click={goToPasswordReset}>Password Reset</button>
    <button on:click={goToAccessControl} disabled>Access Control</button>
</div>
</div>

<style>
    .mainContainer{
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
    }
    h1 {
        font-family: 'poppins';
        font-size: 2rem;
        color: #333;
        text-align: center;
    }
    .options {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 90vh;
        gap: 20px;
        font-family: 'poppins';
    }
    button {
        padding: 10px 20px;
        font-size: 1.2rem;
        cursor: pointer;
        border-radius: 8px;
        border: 0.5px solid #1d1c1c;
        background-color: #5972e4;
        color: white;
    }
    button:hover {
        background-color: #2945e6;
        transition-duration: 0.3s;
        transition: all 0.3s;
    }
</style>
