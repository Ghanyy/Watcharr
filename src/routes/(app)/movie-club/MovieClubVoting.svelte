<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse, MovieClubVoteItem } from "@/types";
	import { voteForMovies, clearVotes } from "@/lib/util/api";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";

	// Component for collapsible reason text
	function CollapsibleReason(reason: string, maxLength: number = 40) {
		return {
			shouldCollapse: reason.length > maxLength,
			truncated: reason.substring(0, maxLength) + "...",
			full: reason,
		};
	}

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{ votesChanged: void }>();

	let submitting = false;
	let selectedVotes: MovieClubVoteItem[] = [];
	let expandedReasons: Set<string> = new Set(); // Track expanded reasons by unique ID

	// Initialize selected votes from current user votes
	$: {
		selectedVotes = cycleData.userVotes.map((vote) => ({
			contentId: vote.contentId,
			priority: vote.priority,
		}));
	}

	$: nominations = cycleData.cycle.nominations || [];
	$: maxVotes = cycleData.maxVotes || 2; // Get from settings, fallback to 2
	$: canVote = cycleData.canVote;

	function toggleVote(contentId: number) {
		const existingIndex = selectedVotes.findIndex(
			(v) => v.contentId === contentId,
		);

		if (existingIndex >= 0) {
			// Remove vote
			const removedPriority = selectedVotes[existingIndex].priority;
			selectedVotes = selectedVotes
				.filter((v) => v.contentId !== contentId)
				.map((v) =>
					v.priority > removedPriority ? { ...v, priority: v.priority - 1 } : v,
				);
		} else if (selectedVotes.length < maxVotes) {
			// Add vote with next priority
			const nextPriority = selectedVotes.length + 1;
			selectedVotes = [...selectedVotes, { contentId, priority: nextPriority }];
		}
	}

	function getVotePriority(contentId: number): number | null {
		const vote = selectedVotes.find((v) => v.contentId === contentId);
		return vote ? vote.priority : null;
	}

	function moveVotePriority(contentId: number, direction: "up" | "down") {
		const currentVote = selectedVotes.find((v) => v.contentId === contentId);
		if (!currentVote) return;

		const newPriority =
			direction === "up" ? currentVote.priority - 1 : currentVote.priority + 1;

		if (newPriority < 1 || newPriority > selectedVotes.length) return;

		// Swap priorities
		const otherVote = selectedVotes.find((v) => v.priority === newPriority);
		if (otherVote) {
			selectedVotes = selectedVotes.map((v) => {
				if (v.contentId === contentId) return { ...v, priority: newPriority };
				if (v.contentId === otherVote.contentId)
					return { ...v, priority: currentVote.priority };
				return v;
			});
		}
	}

	async function submitVotes() {
		if (submitting || selectedVotes.length === 0) return;

		submitting = true;
		const success = await voteForMovies({
			votes: selectedVotes,
			cycleId: cycleData.cycle.id,
		});

		if (success) {
			dispatch("votesChanged");
		}
		submitting = false;
	}

	async function handleClearVotes() {
		if (submitting) return;

		submitting = true;
		const success = await clearVotes(cycleData.cycle.id);

		if (success) {
			selectedVotes = [];
			dispatch("votesChanged");
		}
		submitting = false;
	}

	function getPriorityLabel(priority: number): string {
		switch (priority) {
			case 1:
				return "1st Choice";
			case 2:
				return "2nd Choice";
			case 3:
				return "3rd Choice";
			default:
				return `${priority}th Choice`;
		}
	}

	function toggleReasonExpansion(reasonId: string) {
		if (expandedReasons.has(reasonId)) {
			expandedReasons.delete(reasonId);
		} else {
			expandedReasons.add(reasonId);
		}
		expandedReasons = expandedReasons; // Trigger reactivity
	}
</script>

<div class="voting-section">
	<div class="section-header">
		<h3>
			<Icon icon="check" />
			Vote for Movies
		</h3>
		<p>
			Choose your top {maxVotes} movies from the nominations. Order matters!
		</p>
	</div>

	{#if nominations.length === 0}
		<div class="no-nominations">
			<Icon icon="film" />
			<p>No movies have been nominated yet.</p>
		</div>
	{:else}
		<div class="voting-controls">
			<div class="vote-status">
				<span class="votes-count">
					{selectedVotes.length}/{maxVotes} votes cast
				</span>
				{#if selectedVotes.length > 0}
					<button
						class="clear-votes-btn"
						on:click={handleClearVotes}
						disabled={submitting}
					>
						<Icon icon="trash" />
						Clear Votes
					</button>
				{/if}
			</div>

			{#if selectedVotes.length > 0}
				<button
					class="submit-votes-btn"
					on:click={submitVotes}
					disabled={submitting}
				>
					<Icon icon="check" />
					{submitting ? "Submitting..." : "Submit Votes"}
				</button>
			{/if}
		</div>

		<div class="nominations-grid">
			{#each nominations as nomination}
				{@const priority = getVotePriority(nomination.contentId)}
				{@const isSelected = priority !== null}

				<div
					class="nomination-card"
					class:selected={isSelected}
					on:click={() => toggleVote(nomination.contentId)}
				>
					{#if isSelected}
						<div class="priority-badge">
							{getPriorityLabel(priority)}
						</div>
					{/if}

					<div class="poster-container">
						<Poster
							media={nomination.content}
							showRating={false}
							disableInteraction={true}
						/>
					</div>

					<div class="nomination-details">
						<h5>{nomination.content?.title}</h5>
						<div class="nominators">
							<p class="nominator-label">
								Nominated by:
								{#each nomination.nominators as nominator, i}
									<span class="nominator">{nominator.username}</span
									>{#if i < nomination.nominators.length - 1},
									{/if}
								{/each}
							</p>
						</div>
						{#if nomination.reasons && nomination.reasons.length > 0}
							<div class="reasons">
								{#each nomination.reasons as reason, i}
									{@const reasonData = CollapsibleReason(reason)}
									{@const reasonId = `vote-${nomination.contentId}-${i}`}
									{@const isExpanded = expandedReasons.has(reasonId)}

									<div class="reason-container">
										<p class="reason">
											"{isExpanded
												? reasonData.full
												: reasonData.shouldCollapse
													? reasonData.truncated
													: reasonData.full}"
										</p>
										{#if reasonData.shouldCollapse}
											<button
												class="expand-btn"
												on:click|stopPropagation={() =>
													toggleReasonExpansion(reasonId)}
												type="button"
											>
												{isExpanded ? "Show less" : "Show more"}
											</button>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						{#if isSelected && selectedVotes.length > 1}
							<div class="priority-controls">
								<button
									class="priority-btn"
									class:disabled={priority === 1}
									on:click|stopPropagation={() =>
										moveVotePriority(nomination.contentId, "up")}
									disabled={priority === 1}
								>
									<Icon icon="chevron" />
									Higher
								</button>
								<button
									class="priority-btn"
									class:disabled={priority === selectedVotes.length}
									on:click|stopPropagation={() =>
										moveVotePriority(nomination.contentId, "down")}
									disabled={priority === selectedVotes.length}
								>
									<Icon icon="chevron" />
									Lower
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if selectedVotes.length > 0}
			<div class="vote-summary">
				<h4>Your Vote Summary</h4>
				<div class="vote-list">
					{#each selectedVotes.sort((a, b) => a.priority - b.priority) as vote}
						{@const nomination = nominations.find(
							(n) => n.contentId === vote.contentId,
						)}
						<div class="vote-item">
							<span class="priority">{getPriorityLabel(vote.priority)}</span>
							<span class="title">{nomination?.content?.title}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.voting-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);

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
		--radius-full: 50%;

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

	.section-header {
		text-align: center;
		margin-bottom: var(--space-lg);

		h3 {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: var(--space-sm);
			margin: 0 0 var(--space-sm) 0;
			font-size: 1.5rem;
			font-weight: 700;
			color: var(--text);
			line-height: 1.2;
		}

		p {
			color: var(--text-muted);
			margin: 0;
			line-height: 1.4;
			font-size: 0.95rem;
		}
	}

	.no-nominations {
		text-align: center;
		padding: var(--space-2xl) var(--space-md);
		color: var(--text-muted);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);
		position: relative;
		transition: all 0.2s ease;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--text-muted);
			opacity: 0.02;
			border-radius: inherit;
			pointer-events: none;
		}

		&:hover {
			border-color: var(--primary);

			&::before {
				background: var(--primary);
				opacity: 0.03;
			}
		}

		:global(svg) {
			font-size: 3rem;
			margin-bottom: var(--space-md);
			opacity: 0.4;
			color: var(--text-muted);
		}

		p {
			margin: 0;
			line-height: 1.4;
			position: relative;
			z-index: 1;
		}
	}

	.voting-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
	}

	.vote-status {
		display: flex;
		align-items: center;
		gap: var(--space-md);

		.votes-count {
			font-weight: 600;
			color: var(--primary);
			font-size: 1rem;
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-md);
			border: 2px solid var(--primary);
			position: relative;

			&::before {
				content: "";
				position: absolute;
				inset: 0;
				background: var(--primary);
				opacity: 0.05;
				border-radius: inherit;
				pointer-events: none;
			}
		}

		.clear-votes-btn {
			display: inline-flex;
			align-items: center;
			gap: var(--space-xs);
			padding: var(--space-sm) var(--space-md);
			background: var(--danger, #dc3545);
			color: white;
			border: none;
			border-radius: var(--radius-md);
			font-size: 0.875rem;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);

			&:hover:not(:disabled) {
				background: var(--danger-dark, #c82333);
				transform: translateY(-1px);
				box-shadow: var(--shadow-md);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
				transform: none;
				box-shadow: var(--shadow-sm);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(220, 53, 69, 0.2);
			}

			&:active:not(:disabled) {
				transform: translateY(0);
				box-shadow: var(--shadow-sm);
			}
		}
	}

	.submit-votes-btn {
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
		min-width: 140px;
		justify-content: center;

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

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-lg),
				0 0 0 3px rgba(40, 167, 69, 0.2);
		}

		&:active:not(:disabled) {
			transform: translateY(0);
			box-shadow: var(--shadow-md);
		}
	}

	.nominations-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: var(--space-md);
	}

	.nomination-card {
		background: var(--background);
		border: 2px solid var(--border);
		border-radius: var(--radius-lg);
		overflow: hidden;
		cursor: pointer;
		transition: all 0.3s ease;
		position: relative;
		box-shadow: var(--shadow-sm);

		&:hover {
			transform: translateY(-2px);
			box-shadow: var(--shadow-lg);
			border-color: var(--primary);
		}

		&.selected {
			border-color: var(--primary);
			border-width: 2px;
			box-shadow: var(--shadow-md);
			position: relative;

			&::before {
				content: "";
				position: absolute;
				inset: 0;
				background: var(--primary);
				opacity: 0.04;
				border-radius: inherit;
				pointer-events: none;
			}
		}

		.priority-badge {
			position: absolute;
			top: var(--space-sm);
			right: var(--space-sm);
			background: var(--primary);
			color: white;
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-md);
			font-size: 0.75rem;
			font-weight: 600;
			z-index: 100;
			box-shadow: var(--shadow-md);
			border: 2px solid white;
			text-transform: uppercase;
			letter-spacing: 0.025em;
		}

		.poster-container {
			position: relative;
			aspect-ratio: 2/3;
			overflow: hidden;
			display: flex;
			align-items: center;
			justify-content: center;

			:global(li) {
				list-style: none;
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
				display: flex;
				align-items: center;
				justify-content: center;
				cursor: default !important;
			}

			:global(.container) {
				width: 100% !important;
				height: 100% !important;
				display: flex !important;
				align-items: center !important;
				justify-content: center !important;
				transform: none !important;
				transition: none !important;
			}

			:global(.active .container) {
				transform: none !important;
			}

			:global(.inner) {
				opacity: 0 !important;
				pointer-events: none !important;
			}

			:global(img) {
				max-width: 100%;
				max-height: 100%;
				object-fit: cover;
				object-position: center;
				filter: none !important;
				mix-blend-mode: normal !important;
			}
		}

		.nomination-details {
			padding: var(--space-md);
			position: relative;
			z-index: 1;

			h5 {
				margin: 0 0 var(--space-sm) 0;
				font-size: 0.95rem;
				font-weight: 600;
				line-height: 1.2;
				color: var(--text);
			}

			.nominators {
				margin: 0 0 var(--space-sm) 0;
			}

			.nominator-label {
				font-size: 0.8rem;
				color: var(--text-muted);
				margin: 0;
				font-weight: 500;
				line-height: 1.4;
			}

			.nominator {
				color: var(--primary);
				font-weight: 600;
			}

			.reasons {
				margin: 0 0 var(--space-md) 0;
			}

			.reason-container {
				margin-bottom: var(--space-sm);
				padding: var(--space-sm);
				border-radius: var(--radius-md);
				border-left: 3px solid var(--primary);
				border: 1px solid var(--border);
				position: relative;

				&::before {
					content: "";
					position: absolute;
					inset: 0;
					background: var(--primary);
					opacity: 0.02;
					border-radius: inherit;
					pointer-events: none;
				}
			}

			.reason-container:last-child {
				margin-bottom: var(--space-md);
			}

			.reason {
				font-size: 0.85rem;
				color: var(--text-muted);
				font-style: italic;
				margin: 0 0 var(--space-xs) 0;
				line-height: 1.4;
			}

			.expand-btn {
				background: none;
				border: none;
				color: var(--primary);
				font-size: 0.75rem;
				cursor: pointer;
				padding: var(--space-xs) 0;
				text-decoration: underline;
				font-family: inherit;
				font-weight: 500;
				transition: color 0.2s ease;

				&:hover {
					color: var(--primary-dark, var(--primary));
				}

				&:focus {
					outline: none;
					text-decoration: none;
					box-shadow: 0 0 0 2px var(--primary);
					border-radius: var(--radius-sm);
				}
			}

			.priority-controls {
				display: flex;
				gap: var(--space-sm);
				margin-top: var(--space-sm);

				.priority-btn {
					flex: 1;
					display: inline-flex;
					align-items: center;
					justify-content: center;
					gap: var(--space-xs);
					padding: var(--space-sm);
					background: var(--background-secondary);
					border: 1px solid var(--border);
					border-radius: var(--radius-md);
					font-size: 0.75rem;
					font-weight: 500;
					cursor: pointer;
					transition: all 0.2s ease;
					box-shadow: var(--shadow-sm);

					&:first-child :global(svg) {
						transform: rotate(180deg);
					}

					&:last-child :global(svg) {
						transform: rotate(0deg);
					}

					&:hover:not(:disabled) {
						background: var(--background);
						border-color: var(--primary);
						color: var(--primary);
						box-shadow: var(--shadow-md);
					}

					&:disabled,
					&.disabled {
						opacity: 0.5;
						cursor: not-allowed;
						box-shadow: var(--shadow-sm);
					}

					&:focus {
						outline: none;
						box-shadow:
							var(--shadow-md),
							0 0 0 2px var(--primary);
					}
				}
			}
		}
	}

	.vote-summary {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);
		border-left: 4px solid var(--primary);

		h4 {
			margin: 0 0 var(--space-md) 0;
			color: var(--primary);
			font-weight: 600;
			font-size: 1.125rem;
		}

		.vote-list {
			display: flex;
			flex-direction: column;
			gap: var(--space-sm);
		}

		.vote-item {
			display: flex;
			align-items: center;
			gap: var(--space-md);
			padding: var(--space-sm) var(--space-md);
			border-radius: var(--radius-md);
			border: 1px solid var(--border);
			transition: all 0.2s ease;
			position: relative;

			&::before {
				content: "";
				position: absolute;
				inset: 0;
				background: var(--primary);
				opacity: 0.02;
				border-radius: inherit;
				pointer-events: none;
			}

			&:hover {
				border-color: var(--primary);

				&::before {
					opacity: 0.04;
				}
			}

			.priority {
				font-weight: 600;
				color: white;
				background: var(--primary);
				padding: var(--space-xs) var(--space-sm);
				border-radius: var(--radius-md);
				min-width: 80px;
				text-align: center;
				font-size: 0.8rem;
				text-transform: uppercase;
				letter-spacing: 0.025em;
			}

			.title {
				color: var(--text);
				font-weight: 500;
				flex: 1;
			}
		}
	}

	@media (max-width: 768px) {
		.voting-section {
			gap: var(--space-lg);
		}

		.section-header {
			margin-bottom: var(--space-md);

			h3 {
				font-size: 1.25rem;
			}

			p {
				font-size: 0.9rem;
			}
		}

		.voting-controls {
			flex-direction: column;
			align-items: stretch;
			gap: var(--space-sm);
			padding: var(--space-sm);
		}

		.vote-status {
			justify-content: space-between;
			gap: var(--space-sm);

			.votes-count {
				font-size: 0.9rem;
			}

			.clear-votes-btn {
				padding: var(--space-xs) var(--space-sm);
				font-size: 0.8rem;
			}
		}

		.submit-votes-btn {
			padding: var(--space-xs) var(--space-md);
			font-size: 0.9rem;
			min-width: 120px;
		}

		.nominations-grid {
			grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
			gap: var(--space-sm);
		}

		.nomination-card {
			.priority-badge {
				top: var(--space-xs);
				right: var(--space-xs);
				padding: var(--space-xs);
				font-size: 0.7rem;
			}

			.nomination-details {
				padding: var(--space-sm);

				h5 {
					font-size: 0.85rem;
				}

				.reason,
				.nominator-label {
					font-size: 0.75rem;
				}

				.reason-container {
					padding: var(--space-xs);
				}

				.priority-controls {
					gap: var(--space-xs);

					.priority-btn {
						padding: var(--space-xs);
						font-size: 0.7rem;
					}
				}
			}
		}

		.vote-summary {
			padding: var(--space-md);

			h4 {
				font-size: 1rem;
			}

			.vote-item {
				padding: var(--space-xs) var(--space-sm);
				gap: var(--space-sm);

				.priority {
					min-width: 70px;
					font-size: 0.75rem;
				}

				.title {
					font-size: 0.85rem;
				}
			}
		}
	}
</style>
