<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu from "../Menu.svelte";

	function sortClicked(type: string, modeType: string = "UPDOWN") {
		let mode: string;
		if (modeType === "UPDOWN") {
			mode = "UP";
			if (store.activeSort[0] == type) {
				if (store.activeSort[1] === "UP") {
					mode = "DOWN";
				} else if (store.activeSort[1] === "DOWN") {
					mode = "";
				}
			}
		} else if (modeType === "TOGGLE") {
			mode = "ON";
			if (store.activeSort[0] == type) {
				if (store.activeSort[1] === "ON") {
					mode = "OFF";
				}
			}
		} else {
			console.error("filterClicked() ran without a valid modeType:", modeType);
			return;
		}
		store.activeSort = [type, mode];
	}
</script>

<Menu conf={{ width: "180px", right: "130px", arrowLeft: "21px" }}>
	<button
		class={`plain ${store.activeSort[0] == "DATEADDED" ? store.activeSort[1].toLowerCase() : ""}`}
		onclick={() => sortClicked("DATEADDED")}
	>
		Date Added
	</button>
	<button
		class={`plain ${store.activeSort[0] == "LASTCHANGED" ? store.activeSort[1].toLowerCase() : ""}`}
		onclick={() => sortClicked("LASTCHANGED")}
	>
		Last Changed
	</button>
	<button
		class={`plain ${store.activeSort[0] == "LASTFIN" ? store.activeSort[1].toLowerCase() : ""}`}
		onclick={() => sortClicked("LASTFIN")}
	>
		Last Finished
	</button>
	<button
		class={`plain ${store.activeSort[0] == "RATING" ? store.activeSort[1].toLowerCase() : ""}`}
		onclick={() => sortClicked("RATING")}
	>
		Rating
	</button>
	<button
		class={`plain ${store.activeSort[0] == "ALPHA" ? store.activeSort[1].toLowerCase() : ""}`}
		onclick={() => sortClicked("ALPHA")}
	>
		Alphabetical
	</button>
</Menu>

<style lang="scss">
	button {
		position: relative;

		&.down::before {
			content: "\2193";
		}

		&.up::before {
			content: "\2191";
		}

		&.on::before {
			content: "\2713";
		}

		&::before {
			position: absolute;
			top: 4px;
			left: 12px;
			font-family:
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
			font-size: 18px;
		}
	}

	/* Responsive arrow positioning - using desktop reference */
	:global(.menu) {
		/* Start with desktop value (21px) and scale for mobile */
		@media screen and (max-width: 435px) {
			--al: 19px; /* reduce by 2px for smaller gap */
		}

		@media screen and (max-width: 380px) {
			--al: 17px; /* reduce by another 2px */
		}

		@media screen and (max-width: 375px) {
			--al: 16px; /* reduce by 1px */
		}

		@media screen and (max-width: 370px) {
			--al: 15px; /* reduce by 1px */
		}

		@media screen and (max-width: 350px) {
			--al: 14px; /* reduce by 1px */
		}
	}
</style>
