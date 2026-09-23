package filter

import (
	"sync"
	"wails-app/backend"
)

var (
	ClassifierGenerateEmbedding = backend.GenerateEmbedding
	centroids                   map[string][]float32
	centroidsOnce               sync.Once
)

var dataset = map[string][]string{
	"behavioral": {
		"Tell me about a time you failed.",
		"Describe a situation where you showed leadership.",
		"Give me an example of how you handled a conflict at work.",
		"Walk me through a challenge you overcame.",
		"Tell me about a time you had to learn something new quickly.",
		"Describe a situation where you disagreed with your manager.",
		"Tell me about a time you had to optimize a severely underperforming application.",
		"Describe a situation where you disagreed with a senior engineer about architecture.",
		"Give me an example of a time you led a project under a tight deadline.",
		"Walk me through a time you made a mistake and how you fixed it.",
		"Tell me about a time you had to optimize a severely underperforming React application. What did you do?",
		"Describe a situation where you disagreed with a senior engineer about the architecture of a React component.",
	},
	"coding": {
		"Write a function to reverse a string.",
		"Implement a binary search algorithm.",
		"Write a React hook that debounces a value.",
		"Code a solution for the two-sum problem.",
		"Implement a linked list with insert and delete operations.",
		"Write a function to check if a binary tree is balanced.",
		"Write a custom React hook called useDebounce that takes a value and a delay.",
		"Implement a simple React component that fetches and displays a list of users.",
		"Write a function that flattens a nested array.",
		"Implement rate limiting middleware in Express.",
		"Implement a simple React component that fetches and displays a list of users from an API.",
	},
	"system_design": {
		"How would you design a scalable chat application?",
		"Design the architecture for a URL shortener.",
		"Walk me through designing a distributed caching system.",
		"How would you architect a real-time collaborative editor?",
		"Design a highly available database replication strategy.",
		"How would you design a notification service for millions of users?",
		"How would you design the frontend architecture for a real-time collaborative document editor like Google Docs?",
		"Design a highly scalable e-commerce product listing page with infinite scrolling.",
		"How would you design a global CDN for video streaming?",
		"Walk me through the architecture of a ride-sharing backend.",
		"How would you design the frontend architecture for a real-time collaborative document editor like Google Docs using React?",
	},
	"conceptual": {
		"What is the Virtual DOM?",
		"Explain how HTTP works.",
		"What is the difference between useMemo and useCallback?",
		"How does garbage collection work in JavaScript?",
		"What is eventual consistency in distributed systems?",
		"Explain the difference between SQL and NoSQL databases.",
		"What is the Virtual DOM in React, and how does it improve performance?",
		"Explain the difference between useMemo and useCallback. When should you avoid them?",
		"What are React Server Components and how do they differ from SSR?",
		"What is the event loop in Node.js?",
	},
	"opinion": {
		"Do you prefer React or Angular?",
		"What is your opinion on using TypeScript?",
		"Which state management library do you prefer and why?",
		"What do you think about microservices vs monoliths?",
		"Do you prefer SQL or NoSQL for most projects?",
		"What is your take on test-driven development?",
		"Do you prefer using Redux Toolkit, Zustand, or React Context for global state management?",
		"What is your opinion on using Tailwind CSS versus CSS-in-JS libraries?",
		"Do you think Next.js is strictly better than Vite for modern web apps?",
		"What is your preferred approach to API design: REST or GraphQL?",
		"Do you prefer using Redux Toolkit, Zustand, or React Context for global state management, and why?",
		"What is your opinion on using Tailwind CSS versus CSS-in-JS libraries in a large React project?",
		"What is your opinion on using server-side rendering versus client-side rendering?",
		"What is your opinion on using GraphQL over REST APIs?",
		"What is your opinion on functional programming versus object-oriented programming?",
	},
	"noise": {
		"Okay.",
		"Sounds good, thank you.",
		"Hmm, let me think about that.",
		"Mhm, sure.",
		"Got it.",
		"That's interesting.",
		"Hmm, okay.",
		"That makes sense, thank you.",
		"Sure, I understand.",
		"Let me think for a second.",
	},
	"intro": {
		"Hi, I'm ready for interview.",
		"Let's begin the interview.",
		"I am ready to start.",
		"Hello, let's start.",
		"Ready for my interview.",
		"Start the interview.",
		"Let's get started.",
	},
}

func initCentroids() {
	centroids = make(map[string][]float32)

	for category, examples := range dataset {
		embeddings := make([][]float32, 0, len(examples))
		for _, text := range examples {
			emb := ClassifierGenerateEmbedding(text)
			if len(emb) == 768 {
				embeddings = append(embeddings, emb)
			}
		}

		if len(embeddings) > 0 {
			centroid := make([]float32, 768)
			for _, emb := range embeddings {
				for i := 0; i < 768; i++ {
					centroid[i] += emb[i]
				}
			}
			for i := 0; i < 768; i++ {
				centroid[i] /= float32(len(embeddings))
			}
			centroids[category] = centroid
		}
	}
}

func Classify(embedding []float32) string {
	centroidsOnce.Do(initCentroids)

	if len(embedding) != 768 || len(centroids) == 0 {
		return "unknown"
	}

	bestCategory := "unknown"
	var bestScore float32 = -1.0

	for category, centroid := range centroids {
		score := cosineSimilarity(embedding, centroid)
		if score > bestScore {
			bestScore = score
			bestCategory = category
		}
	}

	return bestCategory
}
