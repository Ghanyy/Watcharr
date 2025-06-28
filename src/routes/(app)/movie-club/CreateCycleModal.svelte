<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import Icon from "@/lib/Icon.svelte";
	import { notify } from "@/lib/util/notify";
	import axios from "axios";

	const dispatch = createEventDispatcher<{
		cycleCreated: void;
		close: void;
	}>();

	let name = "Movie Club Cycle";
	let description = "Choose movies together!";
	let submitting = false;

	async function createCycle() {
		if (submitting || !name.trim()) return;

		submitting = true;
		const nid = notify({ text: "Creating cycle...", type: "loading" });

		try {
			const response = await axios.post("/movie-club/cycle", {
				name: name.trim(),
				description: description.trim()
			});

			notify({ id: nid, text: "Cycle created successfully!", type: "success" });
			
			// Small delay to ensure database consistency
			setTimeout(() => {
				dispatch("cycleCreated");
			}, 100);
		} catch (err: any) {
			console.error("Failed to create cycle:", err);
			const message = err.response?.data?.error || "Failed to create cycle";
			notify({ id: nid, text: message, type: "error" });
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
			createCycle();
		}
	}
</script>

<div class="create-cycle-modal">
	<div class="modal-header">
		<h3>Create New Movie Club Cycle</h3>
		<button class="close-btn" on:click={() => dispatch("close")}>
			<Icon icon="close" />
		</button>
	</div>

	<div class="modal-content">
		<div class="form-group">
			<label for="cycle-name">Cycle Name</label>
			<input
				id="cycle-name"
				type="text"
				bind:value={name}
				placeholder="Movie Club Cycle"
				maxlength="100"
				disabled={submitting}
				on:keydown={handleKeydown}
			/>
		</div>

		<div class="form-group">
			<label for="cycle-description">Description (Optional)</label>
			<textarea
				id="cycle-description"
				bind:value={description}
				placeholder="Describe this cycle..."
				maxlength="500"
				disabled={submitting}
				on:keydown={handleKeydown}
			></textarea>
		</div>

		<div class="cycle-info">
			<h4>Cycle Information</h4>
			<p>
				This cycle will start immediately in the <strong>nomination phase</strong>.
				Users will be able to nominate movies, then vote, and finally see the results.
			</p>
			<p>
				Phase durations and user limits can be configured in the server settings.
			</p>
		</div>

		<div class="action-buttons">
			<button 
				class="cancel-btn" 
				on:click={() => dispatch("close")}
				disabled={submitting}
			>
				Cancel
			</button>
			<button 
				class="create-btn" 
				on:click={createCycle}
				disabled={submitting || !name.trim()}
			>
				{submitting ? "Creating..." : "Create Cycle"}
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.create-cycle-modal {
		width: 100%;
		max-width: 500px;
		background: var(--background);
		border-radius: 12px;
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid var(--border);

		h3 {
			margin: 0;
			color: var(--text);
		}

		.close-btn {
			background: none;
			border: none;
			padding: 0.5rem;
			cursor: pointer;
			color: var(--text-muted);
			border-radius: 4px;

			&:hover {
				color: var(--text);
				background: var(--background-secondary);
			}
		}
	}

	.modal-content {
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;

		label {
			font-weight: 600;
			color: var(--text);
		}

		input, textarea {
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 6px;
			font-family: inherit;
			font-size: 1rem;
			background: var(--background);
			color: var(--text);

			&:focus {
				outline: none;
				border-color: var(--primary);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}

		textarea {
			min-height: 80px;
			resize: vertical;
		}
	}

	.cycle-info {
		background: var(--background-secondary);
		padding: 1rem;
		border-radius: 8px;
		border-left: 3px solid var(--primary);

		h4 {
			margin: 0 0 0.5rem 0;
			color: var(--primary);
		}

		p {
			margin: 0 0 0.5rem 0;
			color: var(--text-muted);
			line-height: 1.4;

			&:last-child {
				margin-bottom: 0;
			}

			strong {
				color: var(--text);
			}
		}
	}

	.action-buttons {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;

		button {
			padding: 0.75rem 1.5rem;
			border-radius: 6px;
			font-size: 1rem;
			cursor: pointer;
			transition: all 0.2s ease;

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}

		.cancel-btn {
			background: var(--background);
			border: 1px solid var(--border);
			color: var(--text);

			&:hover:not(:disabled) {
				background: var(--background-secondary);
				color: var(--text);
			}
		}

		.create-btn {
			background: var(--primary);
			border: none;
			color: white;

			&:hover:not(:disabled) {
				background: var(--primary-dark);
			}
		}
	}

	@media (max-width: 768px) {
		.create-cycle-modal {
			max-width: 95vw;
		}

		.action-buttons {
			flex-direction: column;

			button {
				width: 100%;
			}
		}
	}
</style>