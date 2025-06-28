<script lang="ts">
	import { onMount } from "svelte";
	import { getMovieClubCurrent } from "@/lib/util/api";
	import type { MovieClubCycleResponse } from "@/types";
	import Spinner from "@/lib/Spinner.svelte";
	import MovieClubDashboard from "./MovieClubDashboard.svelte";
	import Error from "@/lib/Error.svelte";

	let cycleData: MovieClubCycleResponse | null = null;
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		try {
			cycleData = await getMovieClubCurrent();
		} catch (err: any) {
			error = err.response?.data?.error || "Failed to load movie club data";
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
				error = err.response?.data?.error || "Failed to load movie club data";
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
			<p>Movie club is either disabled or there's no active cycle running.</p>
			<p>Check back later or ask an admin to start a new cycle!</p>
		</div>
	{:else}
		<MovieClubDashboard {cycleData} on:refresh={refreshData} />
	{/if}
</div>

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

	.refresh-btn {
		background: var(--primary);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		cursor: pointer;
		margin-top: 1rem;

		&:hover {
			background: var(--primary-dark);
		}
	}
</style>