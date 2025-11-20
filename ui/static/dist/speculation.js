if (HTMLScriptElement.supports && HTMLScriptElement.supports('speculationrules')) {
    const specScript = document.createElement('script');
    specScript.type = 'speculationrules';
    specScript.textContent = JSON.stringify({
    prefetch: [
        {
        where: {
            and: [
            { href_matches: '/*' },
            { not: { href_matches: '/logout' } },
            { not: { href_matches: '/signout' } },
            { not: { selector_matches: '[data-no-prefetch]' } }
            ]
        },
        eagerness: 'moderate'
        }
    ],
    prerender: [
        {
        where: {
            href_matches: '/goals'
        },
        eagerness: 'eager'
        }
    ]
    });
    document.head.appendChild(specScript);
}
