<script lang="ts">
	import { onMount } from "svelte";
	import { getMovieClubCurrent } from "@/lib/util/api";
	import type { MovieClubCycleResponse } from "@/types";
	import Spinner from "@/lib/Spinner.svelte";
	import MovieClubDashboard from "./MovieClubDashboard.svelte";
	import Error from "@/lib/Error.svelte";
	import { store } from "@/store.svelte";
	import CreateCycleModal from "./CreateCycleModal.svelte";
	import Modal from "@/lib/Modal.svelte";
	import axios from "axios";
	import { notify } from "@/lib/util/notify";
	import { userHasPermission } from "@/lib/util/helpers";
	import { UserPermission } from "@/types";

	let cycleData: MovieClubCycleResponse | null = null;
	let loading = true;
	let error: string | null = null;
	let showCreateModal = false;
	let movieClubDisabled = false;

	$: isAdmin = store.userInfo && userHasPermission(store.userInfo.permissions, UserPermission.PERM_ADMIN);

	onMount(async () => {
		try {
			cycleData = await getMovieClubCurrent();
		} catch (err: any) {
			if (err.response?.status === 404) {
				const errorMessage = err.response?.data?.error || "";
				console.log("Movie club 404 error message:", errorMessage);
				if (errorMessage.toLowerCase().includes("not enabled") || errorMessage.toLowerCase().includes("disabled")) {
					movieClubDisabled = true;
				} else {
					cycleData = null; // No active cycle
				}
			} else {
				error = err.response?.data?.error || "Failed to load movie club data";
			}
		} finally {
			loading = false;
		}
	});

	async function refreshData() {
		loading = true;
		error = null;
		movieClubDisabled = false;
		try {
			cycleData = await getMovieClubCurrent();
		} catch (err: any) {
			if (err.response?.status === 404) {
				const errorMessage = err.response?.data?.error || "";
				console.log("Movie club 404 error message:", errorMessage);
				if (errorMessage.toLowerCase().includes("not enabled") || errorMessage.toLowerCase().includes("disabled")) {
					movieClubDisabled = true;
				} else {
					cycleData = null; // No active cycle
				}
			} else {
				error = err.response?.data?.error || "Failed to load movie club data";
			}
		} finally {
			loading = false;
		}
	}

	async function deleteCycle() {
		if (!cycleData?.cycle?.id) return;
		
		if (!confirm("Are you sure you want to delete this movie club cycle? This will remove all nominations and votes.")) {
			return;
		}

		const nid = notify({ text: "Deleting cycle...", type: "loading" });
		
		try {
			await axios.delete(`/movie-club/cycle/${cycleData.cycle.id}`);
			notify({ id: nid, text: "Cycle deleted successfully!", type: "success" });
			await refreshData(); // Refresh to show no active cycle
		} catch (err: any) {
			console.error("Failed to delete cycle:", err);
			const message = err.response?.data?.error || "Failed to delete cycle";
			notify({ id: nid, text: message, type: "error" });
		}
	}
</script>

<svelte:head>
	<title>Movie Club - Watcharr</title>
</svelte:head>

<div class="movie-club-page">
	<header>
		<h1>🎬 Movie Club</h1>
		<p>Nominate movies, vote for your favorites, and discover what to watch together!</p>
	</header>

	{#if loading}
		<div class="loading">
			<Spinner />
			<p>Loading movie club...</p>
		</div>
	{:else if error}
		<Error error={error} />
		<button on:click={refreshData} class="refresh-btn">Try Again</button>
	{:else if movieClubDisabled}
		<div class="disabled">
			<h2>Movie Club Disabled</h2>
			<p>The movie club feature is currently disabled.</p>
			{#if isAdmin}
				<p>You can enable it in the server settings.</p>
			{:else}
				<p>Ask an admin to enable it in the server settings.</p>
			{/if}
		</div>
	{:else if !cycleData}
		<div class="no-cycle">
			<h2>No Active Movie Club</h2>
			<p>There's no active movie club cycle running.</p>
			{#if isAdmin}
				<p>As an admin, you can create a new cycle to get started!</p>
				<button class="create-cycle-btn" on:click={() => showCreateModal = true}>
					Create New Cycle
				</button>
			{:else}
				<p>Ask an admin to start a new cycle!</p>
			{/if}
		</div>
	{:else}
		{#if isAdmin}
			<div class="admin-controls">
				<button 
					class="delete-cycle-btn" 
					on:click={deleteCycle}
					disabled={loading}
				>
					Delete Current Cycle
				</button>
			</div>
		{/if}
		<MovieClubDashboard {cycleData} on:refresh={refreshData} />
	{/if}
</div>

{#if showCreateModal}
	<Modal on:close={() => showCreateModal = false}>
		<CreateCycleModal 
			on:cycleCreated={() => {
				showCreateModal = false;
				refreshData();
			}}
			on:close={() => showCreateModal = false}
		/>
	</Modal>
{/if}

<style lang="scss">
	.movie-club-page {
		padding: 2rem 1rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	header {
		text-align: center;
		margin-bottom: 2rem;

		h1 {
			font-size: 2.5rem;
			margin-bottom: 0.5rem;
			color: var(--primary);
		}

		p {
			font-size: 1.1rem;
			color: var(--text-muted);
		}
	}

	.loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		margin: 3rem 0;

		p {
			color: var(--text-muted);
		}
	}

	.no-cycle, .disabled {
		text-align: center;
		margin: 3rem 0;
		padding: 2rem;
		border: 2px dashed var(--border);
		border-radius: 8px;

		h2 {
			margin-bottom: 1rem;
			color: var(--text-muted);
		}

		p {
			color: var(--text-muted);
			margin-bottom: 0.5rem;
		}
	}

	.disabled {
		border-color: var(--danger, #dc3545);
		background: var(--danger-background, rgba(220, 53, 69, 0.1));

		h2 {
			color: var(--danger, #dc3545);
		}
	}

	.refresh-btn, .create-cycle-btn {
		background: var(--primary);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		cursor: pointer;
		margin-top: 1rem;
		font-size: 1rem;

		&:hover {
			background: var(--primary-dark);
		}
	}

	.admin-controls {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 1rem;
		padding: 1rem;
		background: var(--background-secondary);
		border-radius: 8px;
		border-left: 3px solid var(--primary);
	}

	.delete-cycle-btn {
		background: var(--danger, #dc3545);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.9rem;
		transition: all 0.2s ease;

		&:hover:not(:disabled) {
			background: var(--danger-dark, #c82333);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
		}
	}
</style>