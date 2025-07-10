<script lang="ts">
	import { store } from "@/store.svelte";
	import type { WLDetailedViewOption } from "@/types";
	import { page } from "$app/state";
	import Menu from "../Menu.svelte";

	function detailClicked(d: WLDetailedViewOption) {
		if (store.wlDetailedView.includes(d)) {
			store.wlDetailedView = store.wlDetailedView.filter((a) => a !== d);
		} else {
			store.wlDetailedView.push(d);
			store.wlDetailedView = store.wlDetailedView;
		}
	}
</script>

<Menu
	conf={{
		width: "200px",
		right: "132px",
		arrowLeft: page.url?.pathname.startsWith("/search") ? "84px" : "3px",
	}}
>
	<h4 class="norm sm-caps">Shown Details</h4>
	<button
		class={`plain ${store.wlDetailedView?.includes("statusRating") ? "on" : ""}`}
		onclick={() => detailClicked("statusRating")}
	>
		Status & Rating
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("lastWatched") ? "on" : ""}`}
		onclick={() => detailClicked("lastWatched")}
	>
		Watching Season
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("dateAdded") ? "on" : ""}`}
		onclick={() => detailClicked("dateAdded")}
	>
		Date Added
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("dateModified") ? "on" : ""}`}
		onclick={() => detailClicked("dateModified")}
	>
		Date Modified
	</button>
</Menu>

<style lang="scss">
	h4 {
		margin-bottom: 8px;

		&:not(:first-of-type) {
			margin-top: 8px;
		}
	}

	button.plain {
		text-transform: capitalize;
		position: relative;

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
		/* Start with desktop logic and scale appropriately */
		@media screen and (max-width: 435px) {
			--al: 79px; /* search page default - reduce by 5px from 84px */
		}

		@media screen and (max-width: 380px) {
			--al: 74px; /* reduce by another 5px */
		}

		@media screen and (max-width: 375px) {
			--al: 72px; /* reduce by 2px */
		}

		@media screen and (max-width: 370px) {
			--al: 70px; /* reduce by 2px */
		}

		@media screen and (max-width: 350px) {
			--al: 68px; /* reduce by 2px */
		}

		/* Home page uses smaller offset (3px on desktop) */
		:global(body:has([data-sveltekit-preload-code="/"])) & {
			@media screen and (max-width: 435px) {
				--al: 3px;
			}

			@media screen and (max-width: 380px) {
				--al: 3px;
			}

			@media screen and (max-width: 375px) {
				--al: 3px;
			}

			@media screen and (max-width: 370px) {
				--al: 3px;
			}

			@media screen and (max-width: 350px) {
				--al: 3px;
			}
		}
	}
</style>
