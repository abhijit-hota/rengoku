let fileTree = function () {
	return {
		renderLevel: function (obj) {
			const ref = '_' + obj.id;
			let html = `<a href="#" 
                           :class="{'has-children':level.children}" 
                           x-html="' ' + level.name" 
                           ${obj.children ? `@click.prevent="toggleLevel($refs.${ref})"` : ''}>
                        </a>`;

			if (obj.children) {
				html += `<ul style="display:none;" x-ref="${ref}">
                            <template x-for='level in level.children'>
                                <li x-html="renderLevel(level)"></li>
                            </template>
                        </ul>`;
			}
			return html;
		},
		showLevel: function (el) {
			if (el.style.length === 1 && el.style.display === 'none') {
				el.removeAttribute('style');
			} else {
				el.style.removeProperty('display');
			}
			setTimeout(() => {
				el.previousElementSibling.querySelector('i.mdi').classList.add('mdi-folder-open-outline');
				el.previousElementSibling.querySelector('i.mdi').classList.remove('mdi-folder-outline');
				el.classList.add('opacity-100');
			}, 10);
		},
		hideLevel: function (el) {
			el.style.display = 'none';
			el.classList.remove('opacity-100');
			el.previousElementSibling.querySelector('i.mdi').classList.remove('mdi-folder-open-outline');
			el.previousElementSibling.querySelector('i.mdi').classList.add('mdi-folder-outline');

			let refs = el.querySelectorAll('ul[x-ref]');
			for (var i = 0; i < refs.length; i++) {
				this.hideLevel(refs[i]);
			}
		},
		toggleLevel: function (el) {
			if (el.style.length && el.style.display === 'none') {
				this.showLevel(el);
			} else {
				this.hideLevel(el);
			}
		},
	};
};
