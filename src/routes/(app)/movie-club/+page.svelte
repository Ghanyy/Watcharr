<script lang="ts">
	import { onMount } from "svelte";
	import {
		getMovieClubCurrent,
		getActiveMovieClubCycles,
		getMovieClubSettings,
	} from "@/lib/util/api";
	import type { MovieClubCycleResponse, MovieClubSettings } from "@/types";
	import Spinner from "@/lib/Spinner.svelte";
	import MovieClubDashboard from "./MovieClubDashboard.svelte";
	import Error from "@/lib/Error.svelte";
	import { store } from "@/store.svelte";
	import CreateCycleModal from "./CreateCycleModal.svelte";
	import SearchMovieModal from "./SearchMovieModal.svelte";
	import Modal from "@/lib/Modal.svelte";
	import axios from "axios";
	import { notify } from "@/lib/util/notify";
	import { userHasPermission } from "@/lib/util/helpers";
	import { UserPermission } from "@/types";
	import { nominateMovie } from "@/lib/util/api";
	import Icon from "@/lib/Icon.svelte";

	let cycleData: MovieClubCycleResponse | null = null;
	let allActiveCycles: MovieClubCycleResponse[] = [];
	let settings: MovieClubSettings | null = null;
	let loading = true;
	let error: string | null = null;
	let showCreateModal = false;
	let showSearchModal = false;
	let currentNominationCycleId = 0;
	let movieClubDisabled = false;
	let submitting = false;

	$: isAdmin =
		store.userInfo &&
		userHasPermission(store.userInfo.permissions, UserPermission.PERM_ADMIN);

	onMount(async () => {
		try {
			// Fetch both cycle data and settings
			const [cyclesResult, settingsResult] = await Promise.allSettled([
				getActiveMovieClubCycles(),
				getMovieClubSettings(),
			]);

			// Handle cycles data
			if (cyclesResult.status === "fulfilled") {
				allActiveCycles = cyclesResult.value;
				cycleData = allActiveCycles.length > 0 ? allActiveCycles[0] : null;
			} else {
				const err = cyclesResult.reason;
				if (err.response?.status === 404) {
					const errorMessage = err.response?.data?.error || "";
					if (
						errorMessage.toLowerCase().includes("not enabled") ||
						errorMessage.toLowerCase().includes("disabled")
					) {
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
					allActiveCycles = allActiveCycles.map((cycle) => ({
						...cycle,
						maxNominations: settings.nominationsPerUser,
						maxVotes: settings.votesPerUser,
					}));
					if (cycleData) {
						cycleData.maxNominations = settings.nominationsPerUser;
						cycleData.maxVotes = settings.votesPerUser;
					}
				}
			} else {
				console.warn(
					"Failed to load movie club settings:",
					settingsResult.reason,
				);
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
				allActiveCycles = allActiveCycles.map((cycle) => ({
					...cycle,
					maxNominations: settings.nominationsPerUser,
					maxVotes: settings.votesPerUser,
				}));
				if (cycleData) {
					cycleData.maxNominations = settings.nominationsPerUser;
					cycleData.maxVotes = settings.votesPerUser;
				}
			}
		} catch (err: any) {
			if (err.response?.status === 404) {
				const errorMessage = err.response?.data?.error || "";
				if (
					errorMessage.toLowerCase().includes("not enabled") ||
					errorMessage.toLowerCase().includes("disabled")
				) {
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

		if (
			!confirm(
				"Are you sure you want to remove this movie club cycle? This will hide it from the active cycles list but preserve the data.",
			)
		) {
			return;
		}

		const nid = notify({ text: "Removing cycle...", type: "loading" });

		try {
			await axios.delete(`/movie-club/cycle/${targetCycleId}`);
			notify({ id: nid, text: "Cycle removed successfully!", type: "success" });
			await refreshData(); // Refresh to update the cycles list
		} catch (err: any) {
			console.error("Failed to remove cycle:", err);
			const message = err.response?.data?.error || "Failed to remove cycle";
			notify({ id: nid, text: message, type: "error" });
		}
	}

	async function handleNomination(content: any, reason: string) {
		if (submitting) return;

		submitting = true;
		const success = await nominateMovie({
			contentId: content.tmdbId,
			reason: reason,
			cycleId: currentNominationCycleId,
		});

		if (success) {
			await refreshData();
			showSearchModal = false;
		}
		submitting = false;
	}

	function handleOpenSearchModal(cycleId: number) {
		currentNominationCycleId = cycleId;
		showSearchModal = true;
	}
</script>

<svelte:head>
	<title>Movie Club - Watcharr</title>
</svelte:head>

<div class="movie-club-page">
	<header>
		<h1>🎬 Movie Club</h1>
		<p>
			Nominate movies, vote for your favorites, and discover what to watch
			together!
		</p>
	</header>

	{#if loading}
		<div class="loading">
			<Spinner />
			<p>Loading movie club...</p>
		</div>
	{:else if error}
		<Error {error} />
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
				<button
					class="create-cycle-btn"
					on:click={() => (showCreateModal = true)}
				>
					Create New Cycle
				</button>
			{:else}
				<p>Ask an admin to start a new cycle!</p>
			{/if}
		</div>
	{:else}
		<div class="cycles-container">
			{#each allActiveCycles as cycleData, index}
				<div
					class="cycle-section"
					class:watching-phase={cycleData.cycle.phase === "watching"}
				>
					<div class="cycle-header">
						<h3>
							{cycleData.cycle.name ||
								`Movie Club Cycle #${cycleData.cycle.id}`}
						</h3>
						<div class="cycle-meta">
							<span class="phase-badge phase-{cycleData.cycle.phase}">
								{cycleData.cycle.phase === "nomination"
									? "Nominating"
									: cycleData.cycle.phase === "voting"
										? "Voting"
										: "Watching"}
							</span>
							{#if isAdmin}
								<button
									class="delete-cycle-btn-small"
									on:click={() => deleteCycle(cycleData.cycle.id)}
									disabled={loading}
									title="Remove this cycle"
								>
									×
								</button>
							{/if}
						</div>
					</div>
					<MovieClubDashboard
						{cycleData}
						on:refresh={refreshData}
						on:openSearchModal={() => handleOpenSearchModal(cycleData.cycle.id)}
						on:nominateMovie={(e) =>
							handleNomination(e.detail.content, e.detail.reason)}
					/>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Always visible admin controls at bottom -->
	{#if isAdmin && !movieClubDisabled}
		<div class="bottom-admin-controls">
			<button
				class="start-new-cycle-btn"
				on:click={() => (showCreateModal = true)}
				disabled={loading}
				title="Start a new movie club cycle"
			>
				<Icon icon="plus" />
				Start New Cycle
			</button>
		</div>
	{/if}
</div>

{#if showCreateModal}
	<Modal on:close={() => (showCreateModal = false)}>
		<CreateCycleModal
			on:cycleCreated={() => {
				showCreateModal = false;
				refreshData();
			}}
			on:close={() => (showCreateModal = false)}
		/>
	</Modal>
{/if}

{#if showSearchModal}
	<Modal on:close={() => (showSearchModal = false)}>
		<SearchMovieModal
			on:movieSelected={(e) =>
				handleNomination(e.detail.content, e.detail.reason)}
			on:close={() => (showSearchModal = false)}
		/>
	</Modal>
{/if}

<style lang="scss">
	.movie-club-page {
		padding: var(--space-xl) var(--space-md);
		max-width: 1200px;
		margin: 0 auto;

		// CSS custom properties for consistent design system
		--space-xs: 0.25rem;
		--space-sm: 0.5rem;
		--space-md: 1rem;
		--space-lg: 1.5rem;
		--space-xl: 2rem;
		--space-2xl: 3rem;

		--radius-sm: 4px;
		--radius-md: 8px;
		--radius-lg: 12px;
		--radius-xl: 16px;

		--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
		--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
			0 2px 4px -1px rgba(0, 0, 0, 0.06);
		--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
			0 4px 6px -2px rgba(0, 0, 0, 0.05);

		@media (prefers-color-scheme: dark) {
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
			--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4),
				0 2px 4px -1px rgba(0, 0, 0, 0.3);
			--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5),
				0 4px 6px -2px rgba(0, 0, 0, 0.4);
		}
	}

	header {
		text-align: center;
		margin-bottom: var(--space-xl);

		h1 {
			font-size: 2.5rem;
			margin-bottom: var(--space-sm);
			color: var(--primary);
			font-weight: 700;
			line-height: 1.2;
		}

		p {
			font-size: 1.1rem;
			color: var(--text-muted);
			margin: 0;
			line-height: 1.4;
		}
	}

	.loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-md);
		margin: var(--space-2xl) 0;
		padding: var(--space-xl);

		p {
			color: var(--text-muted);
			margin: 0;
		}
	}

	.no-cycle,
	.disabled {
		text-align: center;
		margin: var(--space-2xl) 0;
		padding: var(--space-xl);
		border: 2px dashed var(--border);
		border-radius: var(--radius-lg);
		background: var(--background-secondary);

		h2 {
			margin: 0 0 var(--space-md) 0;
			color: var(--text-muted);
			font-weight: 600;
		}

		p {
			color: var(--text-muted);
			margin: 0 0 var(--space-sm) 0;
			line-height: 1.4;

			&:last-child {
				margin-bottom: 0;
			}
		}
	}

	.disabled {
		border-color: var(--danger, #dc3545);
		background: var(--background-secondary);
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--danger, #dc3545);
			opacity: 0.05;
			border-radius: inherit;
		}

		h2 {
			color: var(--danger, #dc3545);
			position: relative;
			z-index: 1;
		}

		p {
			position: relative;
			z-index: 1;
		}
	}

	.refresh-btn,
	.create-cycle-btn {
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-lg);
		background: var(--primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 500;
		cursor: pointer;
		margin-top: var(--space-md);
		transition: all 0.2s ease;
		box-shadow: var(--shadow-sm);

		&:hover {
			background: var(--primary-dark, var(--primary));
			transform: translateY(-1px);
			box-shadow: var(--shadow-md);
		}

		&:active {
			transform: translateY(0);
			box-shadow: var(--shadow-sm);
		}

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-md),
				0 0 0 3px rgba(59, 130, 246, 0.1);
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
		padding: var(--space-md) 0;
		border-top: 1px solid var(--border);
		margin-top: var(--space-xl);
		display: flex;
		justify-content: center;
		backdrop-filter: blur(10px);
		z-index: 10;
		box-shadow: 0 -4px 6px -1px rgba(0, 0, 0, 0.1);
	}

	.cycles-container {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
		margin-bottom: var(--space-xl);
	}

	.cycle-section {
		background: var(--background);
		border: 2px solid var(--border);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		box-shadow: var(--shadow-lg);
		transition: all 0.3s ease;
		position: relative;
		overflow: hidden;

		&::before {
			content: "";
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			height: 4px;
			background: linear-gradient(
				90deg,
				var(--primary),
				var(--primary-dark, var(--primary))
			);
			border-radius: var(--radius-xl) var(--radius-xl) 0 0;
		}

		&:hover {
			transform: translateY(-2px);
			box-shadow:
				var(--shadow-lg),
				0 8px 25px rgba(0, 0, 0, 0.15);
			border-color: var(--primary);
		}

		&.watching-phase {
			border-color: var(--success, #22c55e);

			&::before {
				background: linear-gradient(90deg, var(--success, #22c55e), #16a34a);
			}

			&:hover {
				border-color: var(--success, #22c55e);
				box-shadow:
					var(--shadow-lg),
					0 8px 25px rgba(34, 197, 94, 0.2);
			}
		}

		& + .cycle-section {
			position: relative;

			&::after {
				content: "";
				position: absolute;
				top: calc(-1 * var(--space-2xl) / 2 - 1px);
				left: 50%;
				transform: translateX(-50%);
				width: 60%;
				height: 2px;
				background: linear-gradient(
					90deg,
					transparent,
					var(--border),
					transparent
				);
			}
		}
	}

	.cycle-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);

		h3 {
			margin: 0;
			font-size: 1.5rem;
			font-weight: 700;
			color: var(--text);
			display: flex;
			align-items: center;
			gap: var(--space-sm);

			&::before {
				content: "🎬";
				font-size: 1.25rem;
			}
		}

		.cycle-meta {
			display: flex;
			align-items: center;
			gap: var(--space-md);
		}

		.phase-badge {
			display: inline-flex;
			align-items: center;
			padding: var(--space-xs) var(--space-md);
			border-radius: var(--radius-full);
			font-size: 0.875rem;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.025em;
			box-shadow: var(--shadow-sm);
			border: 1px solid transparent;

			&.phase-nomination {
				background: var(--info, #3b82f6);
				color: white;
				border-color: #2563eb;
			}

			&.phase-voting {
				background: var(--warning, #f59e0b);
				color: white;
				border-color: #d97706;
			}

			&.phase-watching {
				background: var(--success, #22c55e);
				color: white;
				border-color: #16a34a;
			}
		}

		.delete-cycle-btn-small {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 28px;
			height: 28px;
			background: var(--danger, #dc3545);
			color: white;
			border: none;
			border-radius: var(--radius-full);
			cursor: pointer;
			font-size: 1rem;
			font-weight: 700;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);

			&:hover:not(:disabled) {
				background: var(--danger-dark, #c82333);
				transform: scale(1.1);
				box-shadow: var(--shadow-md);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
				transform: none;
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 2px rgba(220, 53, 69, 0.3);
			}
		}
	}

	.start-new-cycle-btn {
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-lg);
		background: var(--success, #28a745);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		box-shadow: var(--shadow-md);

		&:hover:not(:disabled) {
			background: var(--success-dark, #218838);
			transform: translateY(-2px);
			box-shadow: var(--shadow-lg);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
			transform: none;
			box-shadow: var(--shadow-sm);
		}

		&:active:not(:disabled) {
			transform: translateY(0);
			box-shadow: var(--shadow-md);
		}

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-lg),
				0 0 0 3px rgba(40, 167, 69, 0.2);
		}
	}

	@media (max-width: 768px) {
		.movie-club-page {
			padding: var(--space-lg) var(--space-sm);
		}

		header h1 {
			font-size: 2rem;
		}

		.bottom-admin-controls {
			padding: var(--space-sm) var(--space-md);
		}

		.start-new-cycle-btn {
			font-size: 0.9rem;
			padding: var(--space-sm) var(--space-md);
		}

		.cycles-container {
			gap: var(--space-xl);
		}

		.cycle-section {
			padding: var(--space-lg);
			border-radius: var(--radius-lg);

			&::before {
				border-radius: var(--radius-lg) var(--radius-lg) 0 0;
			}

			& + .cycle-section::after {
				top: calc(-1 * var(--space-xl) / 2 - 1px);
				width: 80%;
			}
		}

		.cycle-header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
			margin-bottom: var(--space-md);

			h3 {
				font-size: 1.25rem;
			}

			.cycle-meta {
				width: 100%;
				justify-content: space-between;
			}

			.phase-badge {
				font-size: 0.8rem;
				padding: var(--space-xs) var(--space-sm);
			}

			.delete-cycle-btn-small {
				width: 24px;
				height: 24px;
				font-size: 0.9rem;
			}
		}
	}
</style>
