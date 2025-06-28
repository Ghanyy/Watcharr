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

	let cycleData: MovieClubCycleResponse | null = null;
	let loading = true;
	let error: string | null = null;
	let showCreateModal = false;

	$: isAdmin = (store.userInfo?.permissions || 0) & 2; // PERM_ADMIN = 2

	onMount(async () => {
		try {
			cycleData = await getMovieClubCurrent();
		} catch (err: any) {
			if (err.response?.status === 404) {
				cycleData = null; // No active cycle
			} else {
				error = err.response?.data?.error || "Failed to load movie club data";
			}
		} finally {
			loading = false;
		}
	});

	function refreshData() {
		loading = true;
		error = null;
		onMount(async () => {
			try {
				cycleData = await getMovieClubCurrent();
			} catch (err: any) {
				if (err.response?.status === 404) {
					cycleData = null; // No active cycle
				} else {
					error = err.response?.data?.error || "Failed to load movie club data";
				}
			} finally {
				loading = false;
			}
		})();
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

	.no-cycle {
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
</style>