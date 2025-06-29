<script lang="ts">
	import { onMount } from "svelte";
	import { getMovieClubCurrent, getActiveMovieClubCycles, getMovieClubSettings } from "@/lib/util/api";
	import type { MovieClubCycleResponse, MovieClubSettings } from "@/types";
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
	import Icon from "@/lib/Icon.svelte";

	let cycleData: MovieClubCycleResponse | null = null;
	let allActiveCycles: MovieClubCycleResponse[] = [];
	let settings: MovieClubSettings | null = null;
	let loading = true;
	let error: string | null = null;
	let showCreateModal = false;
	let movieClubDisabled = false;

	$: isAdmin = store.userInfo && userHasPermission(store.userInfo.permissions, UserPermission.PERM_ADMIN);

	onMount(async () => {
		try {
			// Fetch both cycle data and settings
			const [cyclesResult, settingsResult] = await Promise.allSettled([
				getActiveMovieClubCycles(),
				getMovieClubSettings()
			]);

			// Handle cycles data
			if (cyclesResult.status === "fulfilled") {
				allActiveCycles = cyclesResult.value;
				cycleData = allActiveCycles.length > 0 ? allActiveCycles[0] : null;
			} else {
				const err = cyclesResult.reason;
				if (err.response?.status === 404) {
					const errorMessage = err.response?.data?.error || "";
					if (errorMessage.toLowerCase().includes("not enabled") || errorMessage.toLowerCase().includes("disabled")) {
						movieClubDisabled = true;
					} else {
						allActiveCycles = []; // No active cycles
						cycleData = null;
					}
				} else {
					error = err.response?.data?.error || "Failed to load movie club data";
				}
			}

			// Handle settings
			if (settingsResult.status === "fulfilled") {
				settings = settingsResult.value;
				// Add settings to all cycle data if available
				if (settings) {
					allActiveCycles = allActiveCycles.map(cycle => ({
						...cycle,
						maxNominations: settings.nominationsPerUser,
						maxVotes: settings.votesPerUser
					}));
					if (cycleData) {
						cycleData.maxNominations = settings.nominationsPerUser;
						cycleData.maxVotes = settings.votesPerUser;
					}
				}
			} else {
				console.warn("Failed to load movie club settings:", settingsResult.reason);
			}
		} catch (err: any) {
			error = "Failed to load movie club data";
		} finally {
			loading = false;
		}
	});

	async function refreshData() {
		loading = true;
		error = null;
		movieClubDisabled = false;
		try {
			allActiveCycles = await getActiveMovieClubCycles();
			cycleData = allActiveCycles.length > 0 ? allActiveCycles[0] : null;
			// Re-apply settings if available
			if (settings) {
				allActiveCycles = allActiveCycles.map(cycle => ({
					...cycle,
					maxNominations: settings.nominationsPerUser,
					maxVotes: settings.votesPerUser
				}));
				if (cycleData) {
					cycleData.maxNominations = settings.nominationsPerUser;
					cycleData.maxVotes = settings.votesPerUser;
				}
			}
		} catch (err: any) {
			if (err.response?.status === 404) {
				const errorMessage = err.response?.data?.error || "";
				if (errorMessage.toLowerCase().includes("not enabled") || errorMessage.toLowerCase().includes("disabled")) {
					movieClubDisabled = true;
				} else {
					allActiveCycles = []; // No active cycles
					cycleData = null;
				}
			} else {
				error = err.response?.data?.error || "Failed to load movie club data";
			}
		} finally {
			loading = false;
		}
	}

	async function deleteCycle(cycleId?: number) {
		// Use the provided cycleId or fall back to the first cycle's ID for backwards compatibility
		const targetCycleId = cycleId || cycleData?.cycle?.id;
		if (!targetCycleId) return;
		
		if (!confirm("Are you sure you want to delete this movie club cycle? This will remove all nominations and votes.")) {
			return;
		}

		const nid = notify({ text: "Deleting cycle...", type: "loading" });
		
		try {
			await axios.delete(`/movie-club/cycle/${targetCycleId}`);
			notify({ id: nid, text: "Cycle deleted successfully!", type: "success" });
			await refreshData(); // Refresh to update the cycles list
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
	{:else if allActiveCycles.length === 0}
		<div class="no-cycle">
			<h2>No Active Movie Club</h2>
			<p>There are no active movie club cycles running.</p>
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
		<div class="cycles-container">
			{#each allActiveCycles as cycleData, index}
				<div class="cycle-section" class:watching-phase={cycleData.cycle.phase === 'watching'}>
					<div class="cycle-header">
						<h3>{cycleData.cycle.name || `Movie Club Cycle #${cycleData.cycle.id}`}</h3>
						<div class="cycle-meta">
							<span class="phase-badge phase-{cycleData.cycle.phase}">
								{cycleData.cycle.phase === 'nomination' ? 'Nominating' : 
								 cycleData.cycle.phase === 'voting' ? 'Voting' : 'Watching'}
							</span>
							{#if isAdmin}
								<button 
									class="delete-cycle-btn-small" 
									on:click={() => deleteCycle(cycleData.cycle.id)}
									disabled={loading}
									title="Delete this cycle"
								>
									×
								</button>
							{/if}
						</div>
					</div>
					<MovieClubDashboard {cycleData} on:refresh={refreshData} />
				</div>
			{/each}
		</div>
	{/if}

	<!-- Always visible admin controls at bottom -->
	{#if isAdmin && !movieClubDisabled}
		<div class="bottom-admin-controls">
			<button 
				class="start-new-cycle-btn"
				on:click={() => showCreateModal = true}
				disabled={loading}
				title={cycleData ? "Start a new cycle (will replace current cycle)" : "Start a new movie club cycle"}
			>
				<Icon icon="plus" />
				Start New Cycle
			</button>
		</div>
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

	.bottom-admin-controls {
		position: sticky;
		bottom: 0;
		background: var(--background);
		padding: 1rem 0;
		border-top: 1px solid var(--border);
		margin-top: 2rem;
		display: flex;
		justify-content: center;
		backdrop-filter: blur(10px);
		z-index: 10;
	}

	.start-new-cycle-btn {
		background: var(--success, #28a745);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		font-weight: 600;
		transition: all 0.2s ease;
		box-shadow: 0 2px 8px rgba(40, 167, 69, 0.3);
		display: flex;
		align-items: center;
		gap: 0.5rem;

		&:hover:not(:disabled) {
			background: var(--success-dark, #218838);
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(40, 167, 69, 0.4);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
			transform: none;
			box-shadow: 0 2px 8px rgba(40, 167, 69, 0.3);
		}

		&:active:not(:disabled) {
			transform: translateY(0);
			box-shadow: 0 2px 6px rgba(40, 167, 69, 0.3);
		}
	}

	@media (max-width: 768px) {
		.bottom-admin-controls {
			padding: 0.75rem 1rem;
		}

		.start-new-cycle-btn {
			font-size: 0.9rem;
			padding: 0.625rem 1.25rem;
		}
	}

	.cycles-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.cycle-section {
		background: var(--background-secondary);
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		border: 2px solid transparent;
		transition: all 0.3s ease;
	}

	.cycle-section.watching-phase {
		border-color: var(--success, #28a745);
		box-shadow: 0 4px 16px rgba(40, 167, 69, 0.2);
	}

	.cycle-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border);
	}

	.cycle-header h3 {
		margin: 0;
		font-size: 1.25rem;
		color: var(--text);
	}

	.cycle-meta {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.phase-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.phase-badge.phase-nomination {
		background: var(--info, #17a2b8);
		color: white;
	}

	.phase-badge.phase-voting {
		background: var(--warning, #ffc107);
		color: #212529;
	}

	.phase-badge.phase-watching {
		background: var(--success, #28a745);
		color: white;
	}

	.delete-cycle-btn-small {
		background: var(--danger, #dc3545);
		color: white;
		border: none;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		cursor: pointer;
		font-size: 1rem;
		font-weight: bold;
		line-height: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s ease;

		&:hover:not(:disabled) {
			background: var(--danger-dark, #c82333);
			transform: scale(1.1);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
			transform: none;
		}
	}

	@media (max-width: 768px) {
		.cycle-section {
			padding: 1rem;
		}

		.cycle-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}

		.cycle-meta {
			width: 100%;
			justify-content: space-between;
		}
	}
</style>