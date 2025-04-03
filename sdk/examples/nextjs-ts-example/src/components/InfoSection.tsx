import { About } from './About';
import { FAQ } from './FAQ';

export function InfoSection() {
    return (
        <div className="mt-4 flex flex-col md:flex-row space-y-4 md:space-y-0 md:space-x-4">
            <About />
            <FAQ />
        </div>
    );
}
